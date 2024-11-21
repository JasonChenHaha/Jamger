package jtcp

import (
	jconfig "jamger/config"
	jlog "jamger/log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Handler func(id uint64, pack *Pack)

type Tcp struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[uint16]Handler
}

// ------------------------- outside -------------------------

func NewTcp() *Tcp {
	return &Tcp{handler: make(map[uint16]Handler)}
}

func (tcp *Tcp) RegisterHandler(id uint16, handler Handler) {
	tcp.handler[id] = handler
}

func (tcp *Tcp) Run() {
	addr := jconfig.GetString("tcp.addr")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", addr)
	go tcp.accept(listener)
	if jconfig.GetBool("debug") {
		go tcp.watch()
	}
}

func (tcp *Tcp) Send(id uint64, pack *Pack) {
	obj, ok := tcp.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(pack)
}

// ------------------------- inside -------------------------

func (tcp *Tcp) accept(listener net.Listener) {
	for {
		con, err := listener.Accept()
		if err != nil {
			jlog.Error(err)
			continue
		} else {
			tcp.add(con)
		}
	}
}

func (tcp *Tcp) add(con net.Conn) {
	id := atomic.AddUint64(&tcp.idc, 1)
	ses := newSes(tcp, con, id)
	tcp.ses.Store(id, ses)
	tcp.counter++
	ses.run()
}

func (tcp *Tcp) delete(id uint64) {
	if obj, ok := tcp.ses.Load(id); ok {
		tcp.ses.Delete(id)
		tcp.counter--
		obj.(*Ses).close()
	}
}

func (tcp *Tcp) receive(id uint64, pack *Pack) {
	fu, ok := tcp.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (tcp *Tcp) watch() {
	ticker := time.NewTicker(3 * time.Second)
	for {
		for range ticker.C {
			jlog.Debug("connecting ", tcp.counter)
		}
	}
}
