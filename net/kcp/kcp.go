package jkcp

import (
	"jconfig"
	"jlog"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtaci/kcp-go"
)

type Handler func(id uint64, pack *Pack)

type Kcp struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[uint16]Handler
}

// ------------------------- outside -------------------------

func NewKcp() *Kcp {
	return &Kcp{handler: make(map[uint16]Handler)}
}

func (kc *Kcp) Register(id uint16, handler Handler) {
	kc.handler[id] = handler
}

func (kc *Kcp) Run() {
	addr := jconfig.GetString("kcp.addr")
	listener, err := kcp.ListenWithOptions(addr, nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", addr)
	go kc.accept(listener)
	if jconfig.GetBool("debug") {
		go kc.watch()
	}
}

func (kc *Kcp) Send(id uint64, cmd uint16, data []byte) {
	obj, ok := kc.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(makePack(cmd, data))
}

// ------------------------- inside -------------------------

func (kc *Kcp) accept(listener *kcp.Listener) {
	for {
		con, err := listener.AcceptKCP()
		if err != nil {
			log.Fatal(err)
			continue
		}
		kc.add(con)
	}
}

func (kc *Kcp) add(con *kcp.UDPSession) {
	id := atomic.AddUint64(&kc.idc, 1)
	ses := newSes(kc, con, id)
	kc.ses.Store(id, ses)
	kc.counter++
	ses.run()
}

func (kc *Kcp) delete(id uint64) {
	if obj, ok := kc.ses.Load(id); ok {
		kc.ses.Delete(id)
		kc.counter--
		obj.(*Ses).close()
	}
}

func (kc *Kcp) receive(id uint64, pack *Pack) {
	fu, ok := kc.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (kc *Kcp) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", kc.counter)
	}
}
