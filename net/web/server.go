package jweb

import (
	"jconfig"
	"jlog"
	"jpb"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Handler func(id uint64, pack *Pack)

type WebSvr struct {
	idc      uint64
	ses      sync.Map
	counter  uint64
	handler  map[jpb.CMD]Handler
	upgrader *websocket.Upgrader
}

// ------------------------- outside -------------------------

func NewWebSvr() *WebSvr {
	ws := &WebSvr{handler: make(map[jpb.CMD]Handler)}
	ws.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", ws.accept)
		server := &http.Server{
			Addr:    jconfig.GetString("web.addr"),
			Handler: mux,
		}
		jlog.Info("listen on ", jconfig.GetString("web.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	if jconfig.GetBool("debug") {
		go ws.watch()
	}
	return ws
}

func (ws *WebSvr) Register(cmd jpb.CMD, handler Handler) {
	ws.handler[cmd] = handler
}

func (ws *WebSvr) Send(id uint64, cmd jpb.CMD, data []byte) {
	obj, ok := ws.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(makePack(cmd, data))
}

// ------------------------- inside -------------------------

func (ws *WebSvr) accept(w http.ResponseWriter, r *http.Request) {
	con, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		jlog.Error(err)
	}
	ws.add(con)
}

func (ws *WebSvr) add(con *websocket.Conn) {
	id := atomic.AddUint64(&ws.idc, 1)
	ses := newSes(ws, con, id)
	ws.ses.Store(id, ses)
	ws.counter++
	ses.run()
}

func (ws *WebSvr) delete(id uint64) {
	if obj, ok := ws.ses.Load(id); ok {
		ws.ses.Delete(id)
		ws.counter--
		obj.(*Ses).close()
	}
}

func (ws *WebSvr) receive(id uint64, pack *Pack) {
	fu, ok := ws.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	fu(id, pack)
}

// ------------------------- debug -------------------------

func (ws *WebSvr) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", ws.counter)
	}
}
