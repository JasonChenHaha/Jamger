package jweb

import (
	jconfig "jamger/config"
	jlog "jamger/log"
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
	return &Web{handler: make(map[uint16]Handler)}
}

func (web *Web) RegisterHandler(id uint16, handler Handler) {
	web.handler[id] = handler
}

func (web *Web) Run() {
	web.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	go func() {
		addr := jconfig.Get("web.addr").(string)
		mux := http.NewServeMux()
		mux.HandleFunc("/", web.accept)
		server := &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  time.Duration(jconfig.Get("web.rTimeout").(int)) * time.Second,
			WriteTimeout: time.Duration(jconfig.Get("web.sTimeout").(int)) * time.Second,
		}
		jlog.Info("listen on ", addr)
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	if jconfig.Get("debug").(bool) {
		go web.watch()
	}
}

func (web *Web) Send(id uint64, pack *Pack) {
	obj, ok := web.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(pack)
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
	ses := newSes(web, id)
	web.ses.Store(id, ses)
	web.counter++
	ses.run(con)
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
	ticker := time.NewTicker(3 * time.Second)
	for {
		for range ticker.C {
			jlog.Debug("connecting ", web.counter)
		}
	}
}
