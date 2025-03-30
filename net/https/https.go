package jhttps

import (
	"encoding/json"
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
	Handler map[jpb.CMD]*Handler
}

type Handler struct {
	Fun      func(*jglobal.Pack)
	Template proto.Message
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

// ------------------------- outside -------------------------

func NewHttps() *Https {
	return &Https{}
}

func (o *Https) AsServer() *Https {
	o.Handler = map[jpb.CMD]*Handler{}
	go func() {
		o.mux = http.NewServeMux()
		o.mux.HandleFunc("/", o.receive)
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

func (o *Https) SetCodec(en func(*jglobal.Pack) error, de func(*jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Https) RegisterPattern(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	o.mux.HandleFunc(pattern, handler)
}

func (o *Https) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.Handler[cmd] = &Handler{
		Fun:      fun,
		Template: template,
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
