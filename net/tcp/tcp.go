package jtcp

import (
	jconfig "jamger/config"
	jlog "jamger/log"
	"net"
	"sync"
	"sync/atomic"
)

type Handler func(id uint64, pack Pack)

type Tcp struct {
	addr    string
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[uint16]Handler
}

// ------------------------- outside -------------------------

func NewTcp(addr string) *Tcp {
	return &Tcp{
		addr:    addr,
		handler: make(map[uint16]Handler),
	}
}

func (tcp *Tcp) RegisterHandler(id uint16, handler Handler) {
	tcp.handler[id] = handler
}

func (tcp *Tcp) Run() {
	cfg := jconfig.Get("tcp").(map[string]any)
	addr := cfg["addr"].(string)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		jlog.Panic(err)
	}
	jlog.Info("tcp listen on ", addr)

	go tcp.accept(listener)
}

func (tcp *Tcp) Send(id uint64, pack Pack) {
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
	ses := newSession(tcp, id)
	tcp.ses.Store(id, ses)
	tcp.counter++
	ses.run(con)
}

func (tcp *Tcp) delete(id uint64) {
	obj, ok := tcp.ses.Load(id)
	if ok {
		tcp.ses.Delete(id)
		tcp.counter--
		obj.(*Ses).close()
	}
}

func (tcp *Tcp) receive(id uint64, pack Pack) {
	fu, ok := tcp.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}
