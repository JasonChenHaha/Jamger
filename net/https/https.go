package jhttps

import (
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Https struct {
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

func NewHttps() *Https {
	return &Https{}
}

func (o *Https) AsServer() *Https {
	o.handler = map[jpb.CMD]*Handler{}
	go func() {
		o.mux = http.NewServeMux()
		o.mux.HandleFunc("/", o.receive)
		server := &http.Server{
			Addr:    jconfig.GetString("https.addr"),
			Handler: o.mux,
		}
		jlog.Info("listen on ", jconfig.GetString("http.addr"))
		if err := server.ListenAndServeTLS("", ""); err != nil {
			jlog.Fatal(err)
		}
	}()
	return o
}

func (o *Https) AsClient() *Https {
	return o
}

func (o *Https) SetCodec(en, de func(string, *jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Https) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

// ------------------------- inside -------------------------

func (o *Https) receive(w http.ResponseWriter, r *http.Request) {
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
			jlog.Error("no transfer cmd.")
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
