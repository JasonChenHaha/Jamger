package jhttps

import (
	"encoding/json"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net/http"
	"strings"

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

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

// ------------------------- outside -------------------------

func NewHttps() *Https {
	return &Https{}
}

func (o *Https) AsServer() *Https {
	o.handler = map[jpb.CMD]*Handler{}
	go func() {
		o.mux = http.NewServeMux()
		o.mux.HandleFunc("/", o.receive)
		o.mux.HandleFunc("/image/", o.imageReceive)
		server := &http.Server{
			Addr:    jconfig.GetString("https.addr"),
			Handler: o.mux,
		}
		jlog.Info("listen on ", jconfig.GetString("https.addr"))
		// if err := server.ListenAndServeTLS(jconfig.GetString("https.crt"), jconfig.GetString("https.key")); err != nil {
		// 	jlog.Fatal(err)
		// }
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return o
}

func (o *Https) AsClient() *Https {
	return o
}

func (o *Https) SetCodec(en, de func(*jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Https) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

func (o *Https) Get(url string) (map[string]any, error) {
	rsp, err := http.Get(url)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	res := map[string]any{}
	if err = json.Unmarshal(body, &res); err != nil {
		jlog.Error(err)
		return nil, err
	}
	return res, nil
}

// ------------------------- inside -------------------------

func (o *Https) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{Data: body}
	if err = decoder(pack); err != nil {
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
	if err = encoder(pack); err != nil {
		return
	}
	if _, err = w.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}

func (o *Https) imageReceive(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	pack := &jglobal.Pack{
		Cmd: jpb.CMD_IMAGE_REQ,
		Data: &jpb.ImageReq{
			Uid: jglobal.Atoi[uint32](parts[len(parts)-1]),
		},
	}
	han := o.handler[pack.Cmd]
	if han == nil {
		jlog.Errorf("no cmd(%s)", pack.Cmd)
		return
	}
	han.fun(pack)
	if pack.Data != nil {
		if _, err := w.Write(pack.Data.([]byte)); err != nil {
			jlog.Error(err)
		}
	}
}
