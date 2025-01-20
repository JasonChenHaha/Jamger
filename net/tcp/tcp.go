package jtcp

import (
	"jconfig"
	"jlog"
	"jpb"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
)

type Callback func(id uint64, cmd jpb.CMD, msg any)

type Handler struct {
	cb  Callback
	msg any
}

type Tcp struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]*Handler
}

// ------------------------- outside -------------------------

func NewTcp() *Tcp {
	tcp := &Tcp{
		handler: make(map[jpb.CMD]*Handler),
	}
	listener, err := net.Listen("tcp", jconfig.GetString("tcp.addr"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", jconfig.GetString("tcp.addr"))
	go tcp.accept(listener)
	if jconfig.GetBool("debug") {
		go tcp.watch()
	}
	return tcp
}

func (tcp *Tcp) Register(cmd jpb.CMD, cb Callback, msg any) {
	tcp.handler[cmd] = &Handler{
		cb:  cb,
		msg: msg,
	}
}

func (tcp *Tcp) Send(id uint64, cmd jpb.CMD, msg any) {
	obj, ok := tcp.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	switch data := msg.(type) {
	case []byte:
		obj.(*Ses).send(&Pack{Cmd: cmd, Data: data})
	case proto.Message:
		raw, err := proto.Marshal(data)
		if err != nil {
			jlog.Errorf("%s, %d", err, cmd)
			return
		}
		obj.(*Ses).send(&Pack{Cmd: cmd, Data: raw})
	}
}

// ------------------------- package -------------------------

func (tcp *Tcp) receive(id uint64, pack *Pack) {
	han, ok := tcp.handler[pack.Cmd]
	if ok {
		msg := proto.Clone(han.msg.(proto.Message))
		if err := proto.Unmarshal(pack.Data, msg); err != nil {
			jlog.Warnf("%s, %d", err, pack.Cmd)
			tcp.delete(id)
			return
		}
		han.cb(id, pack.Cmd, msg)
	} else {
		// han, ok = tcp.handler[jpb.CMD_PASS]
		// if !ok {
		// 	jlog.Error("not register pass cmd.")
		// 	tcp.delete(id)
		// 	return
		// }
		// han.cb(id, pack.Cmd, pack.Data)
	}
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

func (tcp *Tcp) add(con net.Conn) uint64 {
	id := atomic.AddUint64(&tcp.idc, 1)
	ses := newSes(tcp, con, id)
	tcp.ses.Store(id, ses)
	tcp.counter++
	ses.run()
	return id
}

func (tcp *Tcp) delete(id uint64) {
	jlog.Debugln("close session", id)
	if obj, ok := tcp.ses.Load(id); ok {
		tcp.ses.Delete(id)
		tcp.counter--
		obj.(*Ses).close()
	}
}

// ------------------------- debug -------------------------

func (tcp *Tcp) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", tcp.counter)
	}
}
