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

	"google.golang.org/grpc"
)

var Rpc *rpc

type rpc struct {
	server     map[int]*jglobal.HashSlice[int, *jnrpc.Rpc] // map[group] = hs[index, client]
	maglev     map[int]*jglobal.Maglev                     // map[group] = maglev
	roundrobin map[int]uint                                // map[group] = cnt
	mutex      sync.RWMutex
}

// ------------------------- inside -------------------------

func Init() {
	Rpc = &rpc{
		server:     map[int]*jglobal.HashSlice[int, *jnrpc.Rpc]{},
		maglev:     map[int]*jglobal.Maglev{},
		roundrobin: map[int]uint{},
	}
	if jconfig.GetBool("debug") {
		go watch()
	}
}

func join(group int, index int, info map[string]any) {
	Rpc.mutex.Lock()
	defer Rpc.mutex.Unlock()
	if _, ok := Rpc.server[group]; !ok {
		Rpc.server[group] = jglobal.NewHashSlice[int, *jnrpc.Rpc]()
	}
	Rpc.server[group].Insert(group, jnrpc.NewRpc().AsClient(info["addr"].(string)))
	Rpc.maglev[group] = jglobal.NewMaglev(Rpc.server[group].KeyValues())
}

func leave(group int, index int, info map[string]any) {
	Rpc.mutex.Lock()
	defer Rpc.mutex.Unlock()
	if Rpc.server[group] != nil {
		hs := Rpc.server[group]
		hs.Del(index)
		if hs.Len() > 0 {
			Rpc.maglev[group] = jglobal.NewMaglev(Rpc.server[group].KeyValues())
		} else {
			delete(Rpc.server, group)
			delete(Rpc.maglev, group)
			delete(Rpc.roundrobin, group)
		}
	}
}

// ------------------------- outside -------------------------

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

// 指定
func GetDirectTarget(group int, index int) *jnrpc.Rpc {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if _, ok := Rpc.server[group]; ok {
		return Rpc.server[group].Get(index)
	}
	return nil
}

// 轮询
func GetRoundRobinTarget(group int) *jnrpc.Rpc {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if hs, ok := Rpc.server[group]; ok {
		index := Rpc.roundrobin[group]
		Rpc.roundrobin[group]++
		return hs.IndexOf(int(index % uint(hs.Len())))
	}
	return nil
}

// 固定哈希
func GetFixHashTarget(group int, key int) *jnrpc.Rpc {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if hs, ok := Rpc.server[group]; ok {
		return hs.IndexOf(key % hs.Len())
	}
	return nil
}

// 一致性哈希
func GetConsistentHashTarget(group int, key int) *jnrpc.Rpc {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if ml, ok := Rpc.maglev[group]; ok {
		return ml.Get(key)
	}
	return nil
}

// ------------------------- debug -------------------------

func watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		for k, v := range Rpc.server {
			jlog.Debugf("server %d -> %d", k, v.Len())
		}
		for k := range Rpc.maglev {
			jlog.Debug("maglev ", k)
		}
		for k, v := range Rpc.roundrobin {
			jlog.Debugf("roundrobin %d -> %d", k, v)
		}
	}
}
