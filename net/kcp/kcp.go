package jkcp

import (
	jconfig "jamger/config"
	jglobal "jamger/global"
	jlog "jamger/log"
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

func (kc *Kcp) RegisterHandler(id uint16, handler Handler) {
	kc.handler[id] = handler
}

func (kc *Kcp) Run() {
	addr := jconfig.Get("kcp.addr").(string)
	listener, err := kcp.ListenWithOptions(addr, nil, jglobal.G_KCP_DATASHARDS, jglobal.G_KCP_PARITYSHARDS)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", addr)
	go kc.accept(listener)
	if jconfig.Get("debug").(bool) {
		go kc.watch()
	}
}

func (kc *Kcp) Send(id uint64, pack *Pack) {
	obj, ok := kc.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(pack)
}

// ------------------------- debug -------------------------

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
	ses := newSes(kc, id)
	kc.ses.Store(id, ses)
	kc.counter++
	ses.run(con)
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
	ticker := time.NewTicker(3 * time.Second)
	for {
		for range ticker.C {
			jlog.Debug("connecting ", kc.counter)
		}
	}
}
