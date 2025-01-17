package jhttp

import (
	"io"
	"jconfig"
	"jlog"
	"jpb"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
)

type Callback func(w http.ResponseWriter, cmd jpb.CMD, msg proto.Message)

type Handler struct {
	cb  Callback
	msg proto.Message
}

type HttpSvr struct {
	handler map[jpb.CMD]*Handler
}

// ------------------------- outside -------------------------

func NewHttpSvr() *HttpSvr {
	hs := &HttpSvr{
		handler: map[jpb.CMD]*Handler{},
	}
	go func() {
		h2s := &http2.Server{}
		mux := http.NewServeMux()
		mux.HandleFunc("/", hs.receive)
		server := &http.Server{
			Addr:    jconfig.GetString("http.addr"),
			Handler: h2c.NewHandler(mux, h2s),
		}
		http2.ConfigureServer(server, h2s)
		jlog.Info("listen on ", jconfig.GetString("http.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return hs
}

func (hs *HttpSvr) Register(cmd jpb.CMD, cb Callback, msg proto.Message) {
	hs.handler[cmd] = &Handler{
		cb:  cb,
		msg: msg,
	}
}

func (hs *HttpSvr) Response(w http.ResponseWriter, cmd jpb.CMD, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return
	}
	raw := SerializePack(&Pack{Cmd: cmd, Data: data})
	_, err = w.Write(raw)
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return
	}
}

// ------------------------- inside -------------------------

func (hs *HttpSvr) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := UnserializeToPack(body)
	han := hs.handler[pack.Cmd]
	if han == nil {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	msg := proto.Clone(han.msg)
	if err = proto.Unmarshal(pack.Data, msg); err != nil {
		jlog.Warnf("%s, %d", err, pack.Cmd)
		return
	}
	han.cb(w, pack.Cmd, msg)
}
