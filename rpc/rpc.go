package jrpc

import (
	"jconfig"
	"jetcd"
	"jglobal"
	"jlog"
	"jnrpc"
	"sync"
	"time"

	"jschedule"
)

type rpc struct {
	server       map[int]*jglobal.HashSlice[int, *jnrpc.Rpc] // map[group] = hs[index, client]
	maglev       map[int]*jglobal.Maglev[*jnrpc.Rpc]         // map[group] = maglev
	backupMaglev *jglobal.Maglev[*jnrpc.Rpc]
	roundrobin   map[int]uint // map[group] = cnt
	mutex        sync.RWMutex
}

var Rpc *rpc

// ------------------------- outside -------------------------

func Init() *rpc {
	Rpc = &rpc{
		server:     map[int]*jglobal.HashSlice[int, *jnrpc.Rpc]{},
		maglev:     map[int]*jglobal.Maglev[*jnrpc.Rpc]{},
		roundrobin: map[int]uint{},
	}
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, Rpc.watch)
	}
	return Rpc
}

func (o *rpc) Connect(group int) {
	jetcd.Etcd.Watch(group, Rpc.join, Rpc.leave)
}

// 全部
func (o *rpc) GetAllTarget(group int) map[int]*jnrpc.Rpc {
	return o.server[group].KeyValues()
}

// 指定
func (o *rpc) GetDirectTarget(group int, index int) *jnrpc.Rpc {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if _, ok := o.server[group]; ok {
		return o.server[group].Get(index)
	}
	return nil
}

// 轮询
func (o *rpc) GetRoundRobinTarget(group int) *jnrpc.Rpc {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if hs, ok := o.server[group]; ok {
		index := o.roundrobin[group]
		o.roundrobin[group]++
		return hs.IndexOf(int(index % uint(hs.Len())))
	}
	return nil
}

// 固定哈希
func (o *rpc) GetFixHashTarget(group int, key uint32) *jnrpc.Rpc {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if hs, ok := o.server[group]; ok {
		return hs.IndexOf(int(key % uint32(hs.Len())))
	}
	return nil
}

// 一致性哈希
func (o *rpc) GetConsistentHashTarget(group int, key uint32) *jnrpc.Rpc {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if ml, ok := o.maglev[group]; ok {
		return ml.Get(key)
	}
	return nil
}

// 获得本group集群上一次建立一致性哈希映射时，key映射到的节点
func (o *rpc) GetLastConsistentHashTarget(key uint32) *jnrpc.Rpc {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if o.backupMaglev != nil {
		return o.backupMaglev.Get(key)
	}
	return nil
}

// ------------------------- inside -------------------------

func (o *rpc) join(group int, index int, info map[string]any) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	jlog.Debugf("this is %d,%d, join %d,%d", jglobal.GROUP, jglobal.INDEX, group, index)
	if _, ok := o.server[group]; !ok {
		o.server[group] = jglobal.NewHashSlice[int, *jnrpc.Rpc]()
	}
	o.server[group].Insert(index, jnrpc.NewRpc().AsClient(info["addr"].(string)))
	if group == jglobal.GROUP && o.maglev[group] != nil {
		o.backupMaglev = o.maglev[group]
	}
	o.maglev[group] = jglobal.NewMaglev(o.server[group].KeyValues())
}

func (o *rpc) leave(group int, index int, info map[string]any) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.server[group] != nil {
		jlog.Debugf("this is %d,%d, leave %d,%d", jglobal.GROUP, jglobal.INDEX, group, index)
		hs := o.server[group]
		hs.Del(index)
		if hs.Len() > 0 {
			o.maglev[group] = jglobal.NewMaglev(o.server[group].KeyValues())
		} else {
			delete(o.server, group)
			delete(o.maglev, group)
			delete(o.roundrobin, group)
		}
	}
}

// ------------------------- debug -------------------------

func (o *rpc) watch(args ...any) {
	for k, v := range o.server {
		jlog.Debugf("server %d -> %d", k, v.Len())
	}
	for k := range o.maglev {
		jlog.Debug("maglev ", k)
	}
	for k, v := range o.roundrobin {
		jlog.Debugf("roundrobin %d -> %d", k, v)
	}
}
