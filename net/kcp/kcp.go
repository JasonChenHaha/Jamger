package jkcp

import (
	"jconfig"
	"jlog"
	"jpb"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtaci/kcp-go"
)

type Handler func(id uint64, pack *Pack)

type KcpSvr struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]Handler
}

// ------------------------- outside -------------------------

func NewKcpSvr() *KcpSvr {
	ks := &KcpSvr{handler: make(map[jpb.CMD]Handler)}
	listener, err := kcp.ListenWithOptions(jconfig.GetString("kcp.addr"), nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", jconfig.GetString("kcp.addr"))
	go ks.accept(listener)
	if jconfig.GetBool("debug") {
		go ks.watch()
	}
	return ks
}

func (ks *KcpSvr) Register(cmd jpb.CMD, handler Handler) {
	ks.handler[cmd] = handler
}

func (ks *KcpSvr) Send(id uint64, cmd jpb.CMD, data []byte) {
	obj, ok := ks.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(&Pack{
		Cmd:  cmd,
		Data: data,
	})
}

// ------------------------- inside -------------------------

func (ks *KcpSvr) accept(listener *kcp.Listener) {
	for {
		con, err := listener.AcceptKCP()
		if err != nil {
			log.Fatal(err)
			continue
		}
		ks.add(con)
	}
}

func (ks *KcpSvr) add(con *kcp.UDPSession) {
	id := atomic.AddUint64(&ks.idc, 1)
	ses := newSes(ks, con, id)
	ks.ses.Store(id, ses)
	ks.counter++
	ses.run()
}

func (ks *KcpSvr) delete(id uint64) {
	if obj, ok := ks.ses.Load(id); ok {
		ks.ses.Delete(id)
		ks.counter--
		obj.(*Ses).close()
	}
}

func (ks *KcpSvr) receive(id uint64, pack *Pack) {
	fu, ok := ks.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (ks *KcpSvr) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", ks.counter)
	}
}
