package jrpc

import (
	"jconfig"
	"jetcd"
	"jglobal"
	"jlog"
	"jnrpc"
	"log"
	"net"
	"sync"
	"time"

	"jschedule"

	"google.golang.org/grpc"
)

type rpc struct {
	server       map[int]*jglobal.HashSlice[int, *jnrpc.Rpc] // map[group] = hs[index, client]
	maglev       map[int]*jglobal.Maglev[*jnrpc.Rpc]         // map[group] = maglev
	backupMaglev *jglobal.Maglev[*jnrpc.Rpc]
	roundrobin   map[int]uint // map[group] = cnt
	mutex        sync.RWMutex
}

var rc *rpc

// ------------------------- outside -------------------------

func Init() *rpc {
	rc = &rpc{
		server:     map[int]*jglobal.HashSlice[int, *jnrpc.Rpc]{},
		maglev:     map[int]*jglobal.Maglev[*jnrpc.Rpc]{},
		roundrobin: map[int]uint{},
	}
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, watch)
	}
	return rc
}

func Server(desc *grpc.ServiceDesc, svr any) {
	lis, err := net.Listen("tcp", jconfig.GetString("grpc.addr"))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	if t, ok := svr.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(desc, svr)
	go func() {
		err = s.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func Connect(group int) {
	jetcd.Watch(group, join, leave)
}

// 全部
func GetAllTarget(group int) []*jnrpc.Rpc {
	return rc.server[group].Values()
}

// 指定
func GetDirectTarget(group int, index int) *jnrpc.Rpc {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	if _, ok := rc.server[group]; ok {
		return rc.server[group].Get(index)
	}
	return nil
}

// 轮询
func GetRoundRobinTarget(group int) *jnrpc.Rpc {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	if hs, ok := rc.server[group]; ok {
		index := rc.roundrobin[group]
		rc.roundrobin[group]++
		return hs.IndexOf(int(index % uint(hs.Len())))
	}
	return nil
}

// 固定哈希
func GetFixHashTarget(group int, key uint32) *jnrpc.Rpc {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	if hs, ok := rc.server[group]; ok {
		return hs.IndexOf(int(key % uint32(hs.Len())))
	}
	return nil
}

// 一致性哈希
func GetConsistentHashTarget(group int, key uint32) *jnrpc.Rpc {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	if ml, ok := rc.maglev[group]; ok {
		return ml.Get(key)
	}
	return nil
}

// 获得本group集群上一次建立一致性哈希映射时，key映射到的节点
func GetLastConsistentHashTarget(key uint32) *jnrpc.Rpc {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	if rc.backupMaglev != nil {
		return rc.backupMaglev.Get(key)
	}
	return nil
}

// ------------------------- inside -------------------------

func join(group int, index int, info map[string]any) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	jlog.Debugf("this is %d,%d, join %d,%d", jglobal.GROUP, jglobal.INDEX, group, index)
	if _, ok := rc.server[group]; !ok {
		rc.server[group] = jglobal.NewHashSlice[int, *jnrpc.Rpc]()
	}
	rc.server[group].Insert(index, jnrpc.NewRpc().AsClient(info["addr"].(string)))
	if group == jglobal.GROUP && rc.maglev[group] != nil {
		rc.backupMaglev = rc.maglev[group]
	}
	rc.maglev[group] = jglobal.NewMaglev(rc.server[group].KeyValues())
}

func leave(group int, index int, info map[string]any) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	if rc.server[group] != nil {
		jlog.Debugf("this is %d,%d, leave %d,%d", jglobal.GROUP, jglobal.INDEX, group, index)
		hs := rc.server[group]
		hs.Del(index)
		if hs.Len() > 0 {
			rc.maglev[group] = jglobal.NewMaglev(rc.server[group].KeyValues())
		} else {
			delete(rc.server, group)
			delete(rc.maglev, group)
			delete(rc.roundrobin, group)
		}
	}
}

// ------------------------- debug -------------------------

func watch() {
	for k, v := range rc.server {
		jlog.Debugf("server %d -> %d", k, v.Len())
	}
	for k := range rc.maglev {
		jlog.Debug("maglev ", k)
	}
	for k, v := range rc.roundrobin {
		jlog.Debugf("roundrobin %d -> %d", k, v)
	}
}
