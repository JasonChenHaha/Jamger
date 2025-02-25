package jetcd

import (
	"context"
	"fmt"
	"jconfig"
	"jglobal"
	"jlog"
	"jschedule"
	"juBase"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Handler func(group int, index int, info map[string]any) // handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件

type etcd struct {
	*clientv3.Client
	lease  *clientv3.LeaseGrantResponse
	server map[int]map[int]map[string]any // map[group][index] = infos
	watch  map[int][2]Handler             // map[group] = [joinHandler, leaveHandler]
}

var etc *etcd

// ------------------------- outside -------------------------

func Init() {
	etc = &etcd{
		server: map[int]map[int]map[string]any{},
		watch:  map[int][2]Handler{},
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{jconfig.GetString("etcd.addr")}})
	if err != nil {
		jlog.Fatal(err)
	}
	etc.Client = cli
	etc.upload()
	jschedule.DoAt(3*time.Second, etc.update)
	jschedule.DoEvery(time.Duration(jconfig.GetInt("etcd.update"))*time.Millisecond, etc.update)
}

func Watch(group int, join Handler, leave Handler) {
	etc.watch[group] = [2]Handler{join, leave}
}

// ------------------------- inside -------------------------

func (etc *etcd) upload() {
	if etc.lease == nil {
		lease, err := etc.Grant(context.Background(), int64(jconfig.GetInt("etcd.keepalive")/1000))
		if err != nil {
			jlog.Fatal(err)
		}
		etc.lease = lease
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
		defer cancel()
		key := fmt.Sprintf("/%d/%d/%d", jglobal.ZONE, jglobal.GROUP, jglobal.INDEX)
		_, err = etc.Put(ctx, key, jglobal.SerializeServerInfo(), clientv3.WithLease(etc.lease.ID))
		if err != nil {
			jlog.Fatal(err)
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
		defer cancel()
		_, err := etc.KeepAliveOnce(ctx, etc.lease.ID)
		if err != nil {
			jlog.Error(err)
			etc.lease = nil
			juBase.CreateProtect()
		}
	}
}

func (etc *etcd) update() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
	defer cancel()
	rsp, err := etc.Get(ctx, fmt.Sprintf("/%d/", jglobal.ZONE), clientv3.WithPrefix())
	if err != nil {
		jlog.Error(err)
	}
	tmp := map[int]map[int]map[string]any{}
	for _, kv := range rsp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		group, index := jglobal.Atoi[int](parts[2]), jglobal.Atoi[int](parts[3])
		if tmp[group] == nil {
			tmp[group] = map[int]map[string]any{}
		}
		info := jglobal.UnserializeServerInfo(kv.Value)
		tmp[group][index] = info
		if etc.server[group] == nil || etc.server[group][index] == nil {
			// join
			if _, ok := etc.watch[group]; ok {
				etc.watch[group][0](group, index, info)
			}
		}
	}
	for group, v := range etc.server {
		for index, info := range v {
			if tmp[group] == nil || tmp[group][index] == nil {
				// leave
				if _, ok := etc.watch[group]; ok {
					etc.watch[group][1](group, index, info)
				}
			}
		}
	}
	etc.server = tmp
	etc.upload()
}
