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

	"jschedule"

	"google.golang.org/protobuf/proto"
)

type Handler struct {
	fun func(*jglobal.Pack)
	msg proto.Message
}

type Tcp struct {
	idc     uint64
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]*Handler
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

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
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, tcp.watch)
	}
	return tcp
}

func (tcp *Tcp) AsClient() *Tcp {
	return tcp
}

func (tcp *Tcp) Encoder(fun func(*jglobal.Pack) error) {
	encoder = fun
}

func (tcp *Tcp) Decoder(fun func(*jglobal.Pack) error) {
	decoder = fun
}

func (tcp *Tcp) Register(cmd jpb.CMD, fun func(*jglobal.Pack), msg proto.Message) {
	tcp.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

// ------------------------- package -------------------------

func (tcp *Tcp) receive(ses *Ses, pack *jglobal.Pack) {
	han := tcp.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err := proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, %d", err, pack.Cmd)
			tcp.delete(ses.id)
			return
		}
		han.fun(pack)
	} else {
		if tcp.handler[jpb.CMD_PROXY] == nil {
			jlog.Error("no proxy cmd.")
			tcp.delete(ses.id)
			return
		}
		tcp.handler[jpb.CMD_PROXY].fun(pack)
	}
	if o, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(o)
		if err != nil {
			jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	ses.send(pack)
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

// ------------------------- debug -------------------------

func (tcp *Tcp) watch() {
	jlog.Debug("connecting ", tcp.counter)
}
