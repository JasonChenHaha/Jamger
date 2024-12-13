package jetcd

import (
	"context"
	"fmt"
	"jconfig"
	"jglobal"
	"jlog"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var etc *etcd

type etcd struct {
	*clientv3.Client
	lease *clientv3.LeaseGrantResponse
}

// ------------------------- outside -------------------------

func Init() {
	etc = &etcd{}
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{jconfig.GetString("etcd.addr")}})
	if err != nil {
		jlog.Fatal(err)
	}
	etc.Client = cli

	etc.lease, err = etc.Grant(context.Background(), int64(jconfig.GetInt("etcd.keepalive")))
	if err != nil {
		jlog.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Second)
	defer cancel()
	key := fmt.Sprintf("/services/%s", jglobal.SVR_NAME)
	_, err = etc.Put(ctx, key, jconfig.GetString("grpc.addr"), clientv3.WithLease(etc.lease.ID))
	if err != nil {
		jlog.Fatal(err)
	}
	jglobal.Schedule.DoEvery(fmt.Sprintf("*/%d * * * * *", jconfig.GetInt("etcd.update")/1000), update)
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
	rsp, err := etc.Get(ctx, "/services", clientv3.WithPrefix())
	if err != nil {
		jlog.Error(err)
	}
	for _, kv := range rsp.Kvs {
		jlog.Debugln(string(kv.Key), string(kv.Value))
	}
}
