package jtcp

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"sync"
	"time"

	"jschedule"

	"google.golang.org/protobuf/proto"
)

type Handler struct {
	fun func(*jglobal.Pack)
	msg proto.Message
}

type Tcp struct {
	tmpSes  sync.Map
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]*Handler
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) (uint32, error)

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

func (tcp *Tcp) Decoder(fun func(*jglobal.Pack) (uint32, error)) {
	decoder = fun
}

func (tcp *Tcp) Register(cmd jpb.CMD, fun func(*jglobal.Pack), msg proto.Message) {
	tcp.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

func (tcp *Tcp) Send(uid uint32, msg proto.Message) {

}

// ------------------------- package -------------------------

func (tcp *Tcp) receive(pack *jglobal.Pack) {
	han := tcp.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err := proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, %d", err, pack.Cmd)
			tcp.delete(id)
			return
		}
		han.fun(pack)
	} else {
		if tcp.handler[jpb.CMD_PROXY] == nil {
			jlog.Error("no proxy cmd.")
			tcp.delete(id)
			return
		}
		tcp.handler[jpb.CMD_PROXY].fun(pack)
	}
	obj, ok := tcp.ses.Load(id)
	if !ok {
		jlog.Errorf("session not found")
		return
	}
	if o, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(o)
		if err != nil {
			jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = tmp
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
	ses := newSes(tcp, con)
	tcp.tmpSes.Store(ses, nil)
	tcp.counter++
	ses.run()
}

func (tcp *Tcp) Bind(ses *Ses) {
	tcp.tmpSes.Delete(ses)
	tcp.ses.Store(ses.id, ses)
}

func (tcp *Tcp) delete(ses *Ses) {
	if _, ok := tcp.ses.Load(ses.id); ok {
		tcp.ses.Delete(ses.id)
		tcp.counter--
		ses.close()
	} else if _, ok := tcp.tmpSes.Load(ses); ok {
		tcp.tmpSes.Delete(ses)
		tcp.counter--
		ses.close()
	}
}

// ------------------------- debug -------------------------

func (tcp *Tcp) watch() {
	jlog.Debug("connecting ", tcp.counter)
}
