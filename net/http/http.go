package jhttp

import (
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Http struct {
	mux     *http.ServeMux
	handler map[jpb.CMD]*Handler
}

type Handler struct {
	fun      func(*jglobal.Pack)
	template proto.Message
}

var encoder func(string, *jglobal.Pack) error
var decoder func(string, *jglobal.Pack) error

// ------------------------- outside -------------------------

func NewHttp() *Http {
	return &Http{}
}

func (htp *Http) AsServer() *Http {
	htp.handler = map[jpb.CMD]*Handler{}
	go func() {
		htp.mux = http.NewServeMux()
		htp.mux.HandleFunc("/", htp.receive)
		server := &http.Server{
			Addr:    jconfig.GetString("http.addr"),
			Handler: htp.mux,
		}
		jlog.Info("listen on ", jconfig.GetString("http.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return htp
}

func (htp *Http) AsClient() *Http {
	return htp
}

func (htp *Http) Encoder(fun func(string, *jglobal.Pack) error) {
	encoder = fun
}

func (htp *Http) Decoder(fun func(string, *jglobal.Pack) error) {
	decoder = fun
}

func (htp *Http) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	htp.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

// ------------------------- inside -------------------------

func (htp *Http) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{Data: body}
	if err = decoder(r.URL.Path, pack); err != nil {
		jlog.Warn(err)
		return
	}
	han := htp.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.template)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = msg
	} else {
		han = htp.handler[jpb.CMD_PROXY]
		if han == nil {
			jlog.Error("no proxy cmd.")
			return
		}
	}
	han.fun(pack)
	if o, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(o)
		if err != nil {
			jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	if err = encoder(r.URL.Path, pack); err != nil {
		jlog.Error(err)
		return
	}
	if _, err = w.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}
