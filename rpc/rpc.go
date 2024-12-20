package jrpc

import (
	"jconfig"
	"jetcd"
	"jglobal"
	"jlog"
	"log"
	"net"
	"reflect"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Rpc *rpc

type rpc struct {
	client map[string]any
	server map[string]*jglobal.HashSlice[string, any]
	maglev map[string]*jglobal.Maglev
	mutex  sync.RWMutex
}

// ------------------------- outside -------------------------

func init() {
	Rpc = &rpc{
		client: map[string]any{},
		server: map[string]*jglobal.HashSlice[string, any]{},
		maglev: map[string]*jglobal.Maglev{},
	}
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

func Connect(group string, client any) {
	jetcd.Watch(group, join, leave)
	Rpc.client[group] = client
}

func GetTarget(group string, server string) any {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if _, ok := Rpc.server[group]; ok {
		return Rpc.server[group].Get(server)
	}
	return nil
}

// 固定哈希
func GetFixHashTarget(group string, key int) any {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if hs, ok := Rpc.server[group]; ok {
		return hs.IndexOf(key % hs.Len())
	}
	return nil
}

// 一致性哈希
func GetConsistentHashTarget(group string, key int) any {
	Rpc.mutex.RLock()
	defer Rpc.mutex.RUnlock()
	if ml, ok := Rpc.maglev[group]; ok {
		server := ml.Get(key)
		jlog.Debug(server)
		return Rpc.server[group].Get(server)
	}
	return nil
}

// ------------------------- inside -------------------------

func join(group string, server string, info string) {
	jlog.Debug("join")
	Rpc.mutex.Lock()
	defer Rpc.mutex.Unlock()
	con, err := grpc.NewClient(info, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		jlog.Fatal(err)
	}
	if _, ok := Rpc.server[group]; !ok {
		Rpc.server[group] = jglobal.NewHashSlice[string, any]()
	}
	fn := reflect.ValueOf(Rpc.client[group])
	arg := []reflect.Value{reflect.ValueOf(con)}
	Rpc.server[group].Insert(server, fn.Call(arg)[0].Interface())
	Rpc.maglev[group] = jglobal.NewMaglev(Rpc.server[group].Keys())
}

func leave(group string, server string, info string) {
	Rpc.mutex.Lock()
	defer Rpc.mutex.Unlock()
	if Rpc.server[group] != nil {
		Rpc.server[group].Del(server)
	}
}
