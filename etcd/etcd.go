package jetcd

import (
	"context"
	"fmt"
	"jconfig"
	"jglobal"
	"jlog"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Handler func(group int, index int, info map[string]any) // handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件

var etc *etcd

type etcd struct {
	*clientv3.Client
	lease      *clientv3.LeaseGrantResponse
	server     map[int]map[int]map[string]any // map[group][index] = infos
	joinWatch  map[int][]Handler              // map[group] = handlers
	leaveWatch map[int][]Handler              // map[group] = handlers
}

// ------------------------- outside -------------------------

func Init() {
	etc = &etcd{
		server:     map[int]map[int]map[string]any{},
		joinWatch:  map[int][]Handler{},
		leaveWatch: map[int][]Handler{},
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
	key := fmt.Sprintf("/%d/%d/%d", jglobal.ZONE, jglobal.GROUP, jglobal.INDEX)
	_, err = etc.Put(ctx, key, jglobal.SerializeServerInfo(), clientv3.WithLease(etc.lease.ID))
	if err != nil {
		jlog.Fatal(err)
	}
	jschedule.DoAt(3*time.Second, update)
	jschedule.DoEvery(time.Duration(jconfig.GetInt("etcd.update"))*time.Millisecond, update)
}

func Watch(group int, join Handler, leave Handler) {
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
	rsp, err := etc.Get(ctx, fmt.Sprintf("/%d/", jglobal.ZONE), clientv3.WithPrefix())
	if err != nil {
		jlog.Error(err)
	}
	tmp := map[int]map[int]map[string]any{}
	join := []any{}
	for _, kv := range rsp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		group, index := jglobal.Atoi[int](parts[2]), jglobal.Atoi[int](parts[3])
		if tmp[group] == nil {
			tmp[group] = map[int]map[string]any{}
		}
		info := jglobal.UnserializeServerInfo(kv.Value)
		tmp[group][index] = info
		if etc.server[group] == nil || etc.server[group][index] == nil {
			join = append(join, group, index, info)
		}
	}
	for i := 0; i < len(join); i += 3 {
		for _, f := range etc.joinWatch[join[i].(int)] {
			go f(join[i].(int), join[i+1].(int), join[i+2].(map[string]any))
		}
	}
	leave := []any{}
	for group, v := range etc.server {
		for server, info := range v {
			if tmp[group] == nil || tmp[group][server] == nil {
				leave = append(leave, group, server, info)
			}
		}
	}
	for i := 0; i < len(leave); i += 3 {
		for _, f := range etc.leaveWatch[leave[i].(int)] {
			go f(leave[i].(int), leave[i+1].(int), leave[i+2].(map[string]any))
		}
	}
	etc.server = tmp
}
