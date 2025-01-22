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
	xKcp "github.com/xtaci/kcp-go"
)

type Handler func(id uint64, pack *Pack)

type Kcp struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]Handler
}

// ------------------------- outside -------------------------

func NewKcp() *Kcp {
	return &Kcp{}
}

func (kcp *Kcp) AsServer() *Kcp {
	kcp.handler = map[jpb.CMD]Handler{}
	listener, err := xKcp.ListenWithOptions(jconfig.GetString("kcp.addr"), nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", jconfig.GetString("kcp.addr"))
	go kcp.accept(listener)
	if jconfig.GetBool("debug") {
		go kcp.watch()
	}
	return kcp
}

func (kcp *Kcp) AsClient() *Kcp {
	return kcp
}

func (kcp *Kcp) Register(cmd jpb.CMD, handler Handler) {
	kcp.handler[cmd] = handler
}

func (kcp *Kcp) Send(id uint64, cmd jpb.CMD, data []byte) {
	obj, ok := kcp.ses.Load(id)
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

func (kcp *Kcp) accept(listener *kcp.Listener) {
	for {
		con, err := listener.AcceptKCP()
		if err != nil {
			log.Fatal(err)
			continue
		}
		kcp.add(con)
	}
}

func (kcp *Kcp) add(con *kcp.UDPSession) {
	id := atomic.AddUint64(&kcp.idc, 1)
	ses := newSes(kcp, con, id)
	kcp.ses.Store(id, ses)
	kcp.counter++
	ses.run()
}

func (kcp *Kcp) delete(id uint64) {
	if obj, ok := kcp.ses.Load(id); ok {
		kcp.ses.Delete(id)
		kcp.counter--
		obj.(*Ses).close()
	}
}

func (kcp *Kcp) receive(id uint64, pack *Pack) {
	fu, ok := kcp.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (kcp *Kcp) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", kcp.counter)
	}
}
