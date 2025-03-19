package jweb

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net/http"
	"sync/atomic"
	"time"

	"jschedule"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	fun      func(*jglobal.Pack)
	template proto.Message
}

type Web struct {
	idc      uint64
	ses      *jglobal.Maps[uint64]
	counter  uint64
	handler  map[jpb.CMD]*Handler
	upgrader *websocket.Upgrader
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

// ------------------------- outside -------------------------

func NewWeb() *Web {
	return &Web{}
}

func (o *Web) AsServer() *Web {
	o.ses = jglobal.NewMaps(uint64(1))
	o.handler = map[jpb.CMD]*Handler{}
	o.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", o.accept)
		server := &http.Server{
			Addr:    jconfig.GetString("web.addr"),
			Handler: mux,
		}
		jlog.Info("listen on ", jconfig.GetString("web.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, o.watch)
	}
	return o
}

func (o *Web) AsClient() *Web {
	return o
}

func (o *Web) SetCodec(en, de func(*jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Web) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

func (o *Web) Send(id uint64, pack *jglobal.Pack) {
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

func (o *Web) Close(id uint64) {
	if ses, ok := o.ses.Load(id); ok {
		o.ses.Delete(id)
		o.counter--
		ses.(*Ses).close()
	}
}

// ------------------------- package -------------------------

func (o *Web) receive(id uint64, pack *jglobal.Pack) {
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
			jlog.Error("no cmd(TRANSFER).")
			o.Close(id)
			return
		}
		o.handler[jpb.CMD_TRANSFER].fun(pack)
	}
}

// ------------------------- inside -------------------------

func (o *Web) accept(w http.ResponseWriter, r *http.Request) {
	con, err := o.upgrader.Upgrade(w, r, nil)
	if err != nil {
		jlog.Error(err)
	}
	o.add(con)
}

func (o *Web) add(con *websocket.Conn) {
	id := atomic.AddUint64(&o.idc, 1)
	ses := newSes(o, con, id)
	o.ses.Store(id, ses)
	o.counter++
	ses.run()
}

// ------------------------- debug -------------------------

func (o *Web) watch(args ...any) {
	jlog.Debug("connecting ", o.counter)
}
