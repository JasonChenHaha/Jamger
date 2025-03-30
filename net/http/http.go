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
	Handler map[jpb.CMD]*Handler
}

type Handler struct {
	Fun      func(*jglobal.Pack)
	Template proto.Message
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

// ------------------------- outside -------------------------

func NewHttp() *Http {
	return &Http{}
}

func (o *Http) AsServer() *Http {
	o.Handler = map[jpb.CMD]*Handler{}
	go func() {
		o.mux = http.NewServeMux()
		o.mux.HandleFunc("/", o.receive)
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

func (o *Http) SetCodec(en, de func(*jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Http) RegisterPattern(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	o.mux.HandleFunc(pattern, handler)
}

func (o *Http) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.Handler[cmd] = &Handler{
		Fun:      fun,
		Template: template,
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
	if err = decoder(pack); err != nil {
		return
	}
	han := o.Handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.Template)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.Fun(pack)
	} else {
		han = o.Handler[jpb.CMD_TRANSFER]
		if han == nil {
			jlog.Error("no cmd(TRANSFER).")
			return
		}
		han.Fun(pack)
	}
	if v, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(v)
		if err != nil {
			jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	if err = encoder(pack); err != nil {
		return
	}
	if _, err = w.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}
