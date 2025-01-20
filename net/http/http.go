package jhttp

import (
	"context"
	"io"
	"jconfig"
	"jlog"
	"jpb"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Func func(context.Context, jpb.CMD, proto.Message)
type ProxyFunc func(context.Context, jpb.CMD, []byte)

type Handler struct {
	fun Func
	msg proto.Message
}

type Http struct {
	handler map[jpb.CMD]*Handler
	proxy   ProxyFunc
}

// ------------------------- outside -------------------------

func NewHttp() *Http {
	return &Http{}
}

func (htp *Http) AsServer() *Http {
	htp.handler = map[jpb.CMD]*Handler{}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", htp.receive)
		server := &http.Server{
			Addr:    jconfig.GetString("http.addr"),
			Handler: mux,
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

func (htp *Http) Register(cmd jpb.CMD, fun Func, msg proto.Message) {
	htp.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

func (htp *Http) SetProxy(proxy ProxyFunc) {
	htp.proxy = proxy
}

func (htp *Http) Response(ctx context.Context, cmd jpb.CMD, data any) {
	pack := &Pack{}
	switch o := data.(type) {
	case proto.Message:
		tmp, err := proto.Marshal(o)
		if err != nil {
			jlog.Errorf("%s, cmd: %d", err, pack.cmd)
			return
		}
		pack.data = tmp
	case []byte:
		pack.data = o
	}
	w := ctx.Value(0).(http.ResponseWriter)
	aesKey := ctx.Value(1).([]byte)
	raw, err := encodeFromPack(pack)
	if err != nil {
		jlog.Error(err)
		return
	}
	if _, err = w.Write(raw); err != nil {
		jlog.Error(err)
	}
}

// ------------------------- inside -------------------------

func (htp *Http) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack, err := decodeToPack(body)
	if err != nil {
		jlog.Warn(err)
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, 0, w)
	ctx = context.WithValue(ctx, 1, pack.aesKey)
	han := htp.handler[pack.cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err = proto.Unmarshal(pack.data, msg); err != nil {
			jlog.Warnf("%s, cmd: %d", err, pack.cmd)
			return
		}
		han.fun(ctx, pack.cmd, msg)
	} else {
		if htp.proxy == nil {
			jlog.Error("not register pass cmd.")
			return
		}
		htp.proxy(ctx, pack.cmd, pack.data)
	}
}
