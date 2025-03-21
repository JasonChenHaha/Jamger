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

func (o *Http) AsServer() *Http {
	o.handler = map[jpb.CMD]*Handler{}
	go func() {
		o.mux = http.NewServeMux()
		o.mux.HandleFunc("/", o.receive)
		o.mux.HandleFunc("/.well-known/pki-validation", o.sshVerification)
		server := &http.Server{
			Addr:    jconfig.GetString("http.addr"),
			Handler: o.mux,
		}
		jlog.Info("listen on ", jconfig.GetString("http.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return o
}

func (o *Http) AsClient() *Http {
	return o
}

func (o *Http) SetCodec(en, de func(string, *jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Http) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

// ------------------------- inside -------------------------

func (o *Http) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{Data: body}
	if err = decoder(r.URL.Path, pack); err != nil {
		return
	}
	han := o.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.template)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.fun(pack)
	} else {
		han = o.handler[jpb.CMD_TRANSFER]
		if han == nil {
			jlog.Error("no cmd(TRANSFER).")
			return
		}
		han.fun(pack)
	}
	if v, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(v)
		if err != nil {
			jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	if err = encoder(r.URL.Path, pack); err != nil {
		return
	}
	if _, err = w.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}

func (o *Http) sshVerification(w http.ResponseWriter, r *http.Request) {
	data := []byte("0DC65D90903F77218EF6C3A318C39CF4B6ACA169E1D7E50C9E6EE01E40849465\ncomodoca.com\nacc473308f41d9c")
	if _, err := w.Write(data); err != nil {
		jlog.Error(err)
	}
}
