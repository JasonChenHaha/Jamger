package jtcp

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
)

type Func func(*jglobal.Pack)

type Handler struct {
	fun Func
	msg proto.Message
}

type Tcp struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]*Handler
	proxy   Func
}

// ------------------------- outside -------------------------

func NewTcp() *Tcp {
	return &Tcp{}
}

func (tcp *Tcp) AsServer() *Tcp {
	tcp.handler = map[jpb.CMD]*Handler{}
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

func (tcp *Tcp) AsClient() *Tcp {
	return tcp
}

func (tcp *Tcp) Register(cmd jpb.CMD, fun Func, msg proto.Message) {
	tcp.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

func (tcp *Tcp) SetProxy(proxy Func) {
	tcp.proxy = proxy
}

// ------------------------- package -------------------------

func (tcp *Tcp) receive(pack *jglobal.Pack) {
	han := tcp.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err := proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, %d", err, pack.Cmd)
			tcp.delete(pack.Id)
			return
		}
		han.fun(pack)
	} else {
		if tcp.proxy == nil {
			jlog.Error("not register pass cmd.")
			tcp.delete(pack.Id)
			return
		}
		tcp.proxy(pack)
	}
	obj, ok := tcp.ses.Load(pack.Id)
	if !ok {
		jlog.Errorf("session not found")
		return
	}
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
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
