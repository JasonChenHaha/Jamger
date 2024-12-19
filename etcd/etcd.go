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

type Handler func(group string, server string, info string) // handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件

var etc *etcd

type etcd struct {
	*clientv3.Client
	lease      *clientv3.LeaseGrantResponse
	server     map[string]map[string]string
	joinWatch  map[string][]Handler
	leaveWatch map[string][]Handler
}

// ------------------------- outside -------------------------

func Init() {
	etc = &etcd{
		server:     map[string]map[string]string{},
		joinWatch:  map[string][]Handler{},
		leaveWatch: map[string][]Handler{},
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{jconfig.GetString("etcd.addr")}})
	if err != nil {
		jlog.Fatal(err)
	}
	etc.Client = cli

	etc.lease, err = etc.Grant(context.Background(), int64(jconfig.GetInt("etcd.keepalive")/1000))
	if err != nil {
		jlog.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
	defer cancel()
	key := fmt.Sprintf("/%s/%s/%s", jglobal.ZONE, jglobal.GROUP, jglobal.SERVER)
	_, err = etc.Put(ctx, key, jconfig.GetString("grpc.addr"), clientv3.WithLease(etc.lease.ID))
	if err != nil {
		jlog.Fatal(err)
	}
	jschedule.DoEvery(fmt.Sprintf("*/%d * * * * *", jconfig.GetInt("etcd.update")/1000), update)
}

func Watch(group string, join Handler, leave Handler) {
	etc.joinWatch[group] = append(etc.joinWatch[group], join)
	etc.leaveWatch[group] = append(etc.leaveWatch[group], leave)
}

// ------------------------- inside -------------------------

func update() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
	defer cancel()
	_, err := etc.KeepAliveOnce(ctx, etc.lease.ID)
	if err != nil {
		jlog.Error(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
	defer cancel()
	rsp, err := etc.Get(ctx, fmt.Sprintf("/%s/", jglobal.ZONE), clientv3.WithPrefix())
	if err != nil {
		jlog.Error(err)
	}
	tmp := map[string]map[string]string{}
	for _, kv := range rsp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		group, server := parts[2], parts[3]
		if tmp[group] == nil {
			tmp[group] = map[string]string{}
		}
		info := string(kv.Value)
		tmp[group][server] = info
		if etc.server[group] == nil || etc.server[group][server] == "" {
			// join
			for _, f := range etc.joinWatch[group] {
				f(group, server, info)
			}
		}
	}
	for group, v := range etc.server {
		for server, info := range v {
			if tmp[group] == nil || tmp[group][server] == "" {
				// leave
				for _, f := range etc.leaveWatch[group] {
					f(group, server, info)
				}
			}
		}
	}
	etc.server = tmp
}
