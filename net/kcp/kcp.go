package jkcp

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"log"
	"sync/atomic"
	"time"

	"jschedule"

	"github.com/xtaci/kcp-go"
	xKcp "github.com/xtaci/kcp-go"
)

type Handler func(id uint64, pack *Pack)

type Kcp struct {
	idc     uint64
	ses     *jglobal.Maps[uint64]
	counter uint64
	handler map[jpb.CMD]Handler
}

// ------------------------- outside -------------------------

func NewKcp() *Kcp {
	return &Kcp{}
}

func (o *Kcp) AsServer() *Kcp {
	o.ses = jglobal.NewMaps(uint64(1))
	o.handler = map[jpb.CMD]Handler{}
	listener, err := xKcp.ListenWithOptions(jconfig.GetString("kcp.addr"), nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", jconfig.GetString("kcp.addr"))
	go o.accept(listener)
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, o.watch)
	}
	return o
}

func (o *Kcp) AsClient() *Kcp {
	return o
}

func (o *Kcp) Register(cmd jpb.CMD, handler Handler) {
	o.handler[cmd] = handler
}

func (o *Kcp) Send(id uint64, cmd jpb.CMD, data []byte) {
	obj, ok := o.ses.Load(id)
	if !ok {
		jlog.Errorf("session(%d) not found", id)
		return
	}
	obj.(*Ses).send(&Pack{
		Cmd:  cmd,
		Data: data,
	})
}

// ------------------------- inside -------------------------

func (o *Kcp) accept(listener *kcp.Listener) {
	for {
		con, err := listener.AcceptKCP()
		if err != nil {
			log.Fatal(err)
			continue
		}
		o.add(con)
	}
}

func (o *Kcp) add(con *kcp.UDPSession) {
	id := atomic.AddUint64(&o.idc, 1)
	ses := newSes(o, con, id)
	o.ses.Store(id, ses)
	o.counter++
	ses.run()
}

func (o *Kcp) delete(id uint64) {
	if obj, ok := o.ses.Load(id); ok {
		o.ses.Delete(id)
		o.counter--
		obj.(*Ses).close()
	}
}

func (o *Kcp) receive(id uint64, pack *Pack) {
	fu, ok := o.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (o *Kcp) watch(args ...any) {
	jlog.Debug("connecting ", o.counter)
}
