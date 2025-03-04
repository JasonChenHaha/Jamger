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
)

type Handler func(id uint64, pack *Pack)

type Web struct {
	idc      uint64
	ses      *jglobal.Maps[uint64]
	counter  uint64
	handler  map[jpb.CMD]Handler
	upgrader *websocket.Upgrader
}

// ------------------------- outside -------------------------

func NewWeb() *Web {
	return &Web{}
}

func (o *Web) AsServer() *Web {
	o.ses = jglobal.NewMaps(uint64(1))
	o.handler = map[jpb.CMD]Handler{}
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

func (o *Web) Register(cmd jpb.CMD, handler Handler) {
	o.handler[cmd] = handler
}

func (o *Web) Send(id uint64, cmd jpb.CMD, data []byte) {
	obj, ok := o.ses.Load(id)
	if !ok {
		jlog.Errorf("session(%d) not found", id)
		return
	}
	obj.(*Ses).send(makePack(cmd, data))
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

func (o *Web) delete(id uint64) {
	if obj, ok := o.ses.Load(id); ok {
		o.ses.Delete(id)
		o.counter--
		obj.(*Ses).close()
	}
}

func (o *Web) receive(id uint64, pack *Pack) {
	fu, ok := o.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (o *Web) watch(args ...any) {
	jlog.Debug("connecting ", o.counter)
}
