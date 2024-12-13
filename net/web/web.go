package jweb

import (
	"jconfig"
	"jlog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Handler func(id uint64, pack *Pack)

type Web struct {
	idc      uint64
	ses      sync.Map
	counter  uint64
	handler  map[uint16]Handler
	upgrader *websocket.Upgrader
}

// ------------------------- outside -------------------------

func NewWeb() *Web {
	web := &Web{handler: make(map[uint16]Handler)}
	web.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	go func() {
		addr := jconfig.GetString("web.addr")
		mux := http.NewServeMux()
		mux.HandleFunc("/", web.accept)
		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		jlog.Info("listen on ", addr)
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	if jconfig.GetBool("debug") {
		go web.watch()
	}
	return web
}

func (web *Web) Register(id uint16, handler Handler) {
	web.handler[id] = handler
}

func (web *Web) Send(id uint64, cmd uint16, data []byte) {
	obj, ok := web.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(makePack(cmd, data))
}

// ------------------------- inside -------------------------

func (web *Web) accept(w http.ResponseWriter, r *http.Request) {
	con, err := web.upgrader.Upgrade(w, r, nil)
	if err != nil {
		jlog.Error(err)
	}
	web.add(con)
}

func (web *Web) add(con *websocket.Conn) {
	id := atomic.AddUint64(&web.idc, 1)
	ses := newSes(web, con, id)
	web.ses.Store(id, ses)
	web.counter++
	ses.run()
}

func (web *Web) delete(id uint64) {
	if obj, ok := web.ses.Load(id); ok {
		web.ses.Delete(id)
		web.counter--
		obj.(*Ses).close()
	}
}

func (web *Web) receive(id uint64, pack *Pack) {
	fu, ok := web.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (web *Web) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", web.counter)
	}
}
