package jetcd

import (
	"context"
	"fmt"
	"jconfig"
	"jglobal"
	"jlog"
	"jschedule"
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

var Etcd *etcd

// ------------------------- outside -------------------------

func Init() {
	Etcd = &etcd{
		server: map[int]map[int]map[string]any{},
		watch:  map[int][2]Handler{},
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{jconfig.GetString("etcd.addr")}})
	if err != nil {
		jlog.Fatal(err)
	}
	Etcd.Client = cli
	jschedule.DoAt(time.Second, Etcd.upload)
	jschedule.DoAt(3*time.Second, Etcd.update)
	jschedule.DoEvery(time.Duration(jconfig.GetInt("etcd.update"))*time.Millisecond, Etcd.update)
}

func (o *etcd) Watch(group int, join Handler, leave Handler) {
	o.watch[group] = [2]Handler{join, leave}
}

// ------------------------- inside -------------------------

func (o *etcd) upload() {
begin:
	if o.lease == nil {
		group, index, info := jglobal.GROUP, jglobal.INDEX, map[string]any{"addr": ""}
		if o.server[group] == nil {
			o.server[group] = map[int]map[string]any{}
		}
		o.server[group][index] = info
		o.watch[jglobal.GROUP][0](group, index, info)
		lease, err := o.Grant(context.Background(), int64(jconfig.GetInt("etcd.keepalive")/1000))
		if err != nil {
			jlog.Fatal(err)
		}
		o.lease = lease
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
		defer cancel()
		key := fmt.Sprintf("/%d/%d/%d", jglobal.ZONE, jglobal.GROUP, jglobal.INDEX)
		_, err = o.Put(ctx, key, jglobal.SerializeServerInfo(), clientv3.WithLease(o.lease.ID))
		if err != nil {
			jlog.Fatal(err)
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
		defer cancel()
		_, err := o.KeepAliveOnce(ctx, o.lease.ID)
		if err != nil {
			jlog.Error(err)
			if err.Error() == "etcdserver: requested lease not found" {
				o.lease = nil
				goto begin
			}
		}
	}
}

func (o *etcd) update() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
	defer cancel()
	rsp, err := o.Get(ctx, fmt.Sprintf("/%d/", jglobal.ZONE), clientv3.WithPrefix())
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
		if o.server[group] == nil || o.server[group][index] == nil {
			// join
			if _, ok := o.watch[group]; ok {
				o.watch[group][0](group, index, info)
			}
		}
	}
	for group, v := range o.server {
		for index, info := range v {
			if tmp[group] == nil || tmp[group][index] == nil {
				// leave
				if _, ok := o.watch[group]; ok {
					o.watch[group][1](group, index, info)
				}
			}
		}
	}
	o.server = tmp
	o.upload()
}
