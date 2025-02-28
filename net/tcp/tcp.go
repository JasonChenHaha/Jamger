package jtcp

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"juser"
	"net"
	"sync"
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
	ses     sync.Map
	counter uint64
	handler map[jpb.CMD]*Handler
}

var encoder func(*jglobal.Pack) error
var decoder func(uint64, *jglobal.Pack) error

// ------------------------- outside -------------------------

func NewTcp() *Tcp {
	return &Tcp{}
}

func (o *Tcp) AsServer() *Tcp {
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

func (o *Tcp) Decoder(fun func(uint64, *jglobal.Pack) error) {
	decoder = fun
}

func (o *Tcp) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

func (o *Tcp) Send(pack *jglobal.Pack) {
	id := pack.Ctx.(jglobal.SesIder).GetSesId()
	if v, ok := o.ses.Load(id); !ok {
		jlog.Errorf("no session(%d)", id)
		return
	} else {
		if v, ok := pack.Data.(proto.Message); ok {
			tmp, err := proto.Marshal(v)
			if err != nil {
				jlog.Errorf("%s, cmd(%d)", err, pack.Cmd)
				return
			}
			pack.Data = tmp
		}
		jlog.Debugf("tcp send to C %d, cmd(%d), data(%v)", pack.Ctx.(*juser.User).Uid, pack.Cmd, pack.Data)
		v.(*Ses).send(pack)
	}
}

// ------------------------- package -------------------------

func (o *Tcp) receive(ses *Ses, pack *jglobal.Pack) {
	han := o.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.template)
		if err := proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%d)", err, pack.Cmd)
			o.delete(ses.id)
			return
		}
		pack.Data = msg
		pack.Ctx.(jglobal.Locker).Lock()
		han.fun(pack)
		pack.Ctx.(jglobal.Locker).UnLock()
	} else {
		if o.handler[jpb.CMD_PROXY] == nil {
			jlog.Error("no proxy cmd.")
			o.delete(ses.id)
			return
		}
		o.handler[jpb.CMD_PROXY].fun(pack)
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

func (o *Tcp) delete(id uint64) {
	if obj, ok := o.ses.Load(id); ok {
		o.ses.Delete(id)
		o.counter--
		obj.(*Ses).close()
	}
}

// ------------------------- debug -------------------------

func (o *Tcp) watch(args ...any) {
	jlog.Debug("connecting ", o.counter)
}
