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

type TcpSvr struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]*Handler
}

// ------------------------- outside -------------------------

func NewTcpSvr() *TcpSvr {
	ts := &TcpSvr{
		handler: make(map[jpb.CMD]*Handler),
	}
	listener, err := net.Listen("tcp", jconfig.GetString("tcp.addr"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", jconfig.GetString("tcp.addr"))
	go ts.accept(listener)
	if jconfig.GetBool("debug") {
		go ts.watch()
	}
	return ts
}

func (ts *TcpSvr) Register(cmd jpb.CMD, cb Callback, msg any) {
	ts.handler[cmd] = &Handler{
		cb:  cb,
		msg: msg,
	}
}

func (ts *TcpSvr) Send(id uint64, cmd jpb.CMD, msg any) {
	obj, ok := ts.ses.Load(id)
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

func (ts *TcpSvr) receive(id uint64, pack *Pack) {
	han, ok := ts.handler[pack.Cmd]
	if ok {
		msg := proto.Clone(han.msg.(proto.Message))
		if err := proto.Unmarshal(pack.Data, msg); err != nil {
			jlog.Warnf("%s, %d", err, pack.Cmd)
			ts.delete(id)
			return
		}
		han.cb(id, pack.Cmd, msg)
	} else {
		han, ok = ts.handler[jpb.CMD_PASS]
		if !ok {
			jlog.Error("not register pass cmd.")
			ts.delete(id)
			return
		}
		han.cb(id, pack.Cmd, pack.Data)
	}
}

// ------------------------- inside -------------------------

func (ts *TcpSvr) accept(listener net.Listener) {
	for {
		con, err := listener.Accept()
		if err != nil {
			jlog.Error(err)
			continue
		} else {
			ts.add(con)
		}
	}
}

func (ts *TcpSvr) add(con net.Conn) uint64 {
	id := atomic.AddUint64(&ts.idc, 1)
	ses := newSes(ts, con, id)
	ts.ses.Store(id, ses)
	ts.counter++
	ses.run()
	return id
}

func (ts *TcpSvr) delete(id uint64) {
	jlog.Debugln("close session", id)
	if obj, ok := ts.ses.Load(id); ok {
		ts.ses.Delete(id)
		ts.counter--
		obj.(*Ses).close()
	}
}

// ------------------------- debug -------------------------

func (ts *TcpSvr) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", ts.counter)
	}
}
