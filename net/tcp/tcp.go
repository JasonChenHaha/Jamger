package jtcp

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"sync/atomic"
	"time"

	"jschedule"

	"google.golang.org/protobuf/proto"
)

type Handler struct {
	fun      func(*jglobal.Pack)
	template proto.Message
}

type Tcp struct {
	idc     uint64
	ses     *jglobal.Maps[uint64]
	counter uint64
	handler map[jpb.CMD]*Handler
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

// ------------------------- outside -------------------------

func NewTcp() *Tcp {
	return &Tcp{}
}

func (o *Tcp) AsServer() *Tcp {
	o.ses = jglobal.NewMaps(uint64(1))
	o.handler = map[jpb.CMD]*Handler{}
	listener, err := net.Listen("tcp", jconfig.GetString("tcp.addr"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", jconfig.GetString("tcp.addr"))
	go o.accept(listener)
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, o.watch)
	}
	return o
}

func (o *Tcp) AsClient() *Tcp {
	return o
}

func (o *Tcp) Encoder(fun func(*jglobal.Pack) error) {
	encoder = fun
}

func (o *Tcp) Decoder(fun func(*jglobal.Pack) error) {
	decoder = fun
}

func (o *Tcp) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

func (o *Tcp) Send(pack *jglobal.Pack) {
	id := pack.Ctx.(jglobal.User1).GetSesId()
	ses, ok := o.ses.Load(id)
	if !ok {
		jlog.Errorf("no session(%d)", id)
		return
	}
	if v, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(v)
		if err != nil {
			jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	ses.(*Ses).send(pack)
}

func (o *Tcp) Close(id uint64) {
	if obj, ok := o.ses.Load(id); ok {
		o.ses.Delete(id)
		o.counter--
		obj.(*Ses).close()
	}
}

// ------------------------- package -------------------------

func (o *Tcp) receive(id uint64, pack *jglobal.Pack) {
	han := o.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.template)
		if err := proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			o.Close(id)
			return
		}
		pack.Data = msg
		han.fun(pack)
	} else {
		if o.handler[jpb.CMD_TRANSFER] == nil {
			jlog.Error("no transfer cmd.")
			o.Close(id)
			return
		}
		o.handler[jpb.CMD_TRANSFER].fun(pack)
	}
}

// ------------------------- inside -------------------------

func (o *Tcp) accept(listener net.Listener) {
	for {
		con, err := listener.Accept()
		if err != nil {
			jlog.Error(err)
			continue
		} else {
			o.add(con)
		}
	}
}

func (o *Tcp) add(con net.Conn) {
	id := atomic.AddUint64(&o.idc, 1)
	ses := newSes(o, con, id)
	o.ses.Store(id, ses)
	o.counter++
	ses.run()
}

// ------------------------- debug -------------------------

func (o *Tcp) watch(args ...any) {
	jlog.Debug("connecting ", o.counter)
}
