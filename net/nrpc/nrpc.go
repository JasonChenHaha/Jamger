package jnrpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"jconfig"
	"jlog"
	"jpb"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
)

type Func func(w http.ResponseWriter, cmd jpb.CMD, msg proto.Message)

type Handler struct {
	fun Func
	msg proto.Message
}

type Rpc struct {
	addr    string
	client  *http.Client
	handler map[jpb.CMD]*Handler
}

// ------------------------- outside -------------------------

func NewRpc() *Rpc {
	return &Rpc{}
}

func (rpc *Rpc) AsServer() *Rpc {
	rpc.handler = map[jpb.CMD]*Handler{}
	go func() {
		h2s := &http2.Server{}
		mux := http.NewServeMux()
		mux.HandleFunc("/", rpc.receive)
		server := &http.Server{
			Addr:    jconfig.GetString("rpc.addr"),
			Handler: h2c.NewHandler(mux, h2s),
		}
		http2.ConfigureServer(server, h2s)
		jlog.Info("listen on ", jconfig.GetString("rpc.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return rpc
}

func (rpc *Rpc) AsClient(addr string) *Rpc {
	rpc.addr = "http://" + addr
	rpc.client = &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
		Timeout: time.Duration(jconfig.GetInt("rpc.timeout")) * time.Millisecond,
	}
	return rpc
}

func (rpc *Rpc) Register(cmd jpb.CMD, fun Func, msg proto.Message) {
	rpc.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

func (rpc *Rpc) Proxy(cmd jpb.CMD, data []byte) (jpb.CMD, []byte) {
	raw := encodeFromPack(&Pack{cmd: cmd, data: data})
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(raw))
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return 0, nil
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return 0, nil
	}
	if len(body) == 0 {
		jlog.Errorf("body is empty")
		return 0, nil
	}
	pack := decodeToPack(body)
	return pack.cmd, pack.data
}

func (rpc *Rpc) Call(cmd jpb.CMD, in proto.Message, out proto.Message) (jpb.CMD, proto.Message) {
	data, err := proto.Marshal(in)
	if err != nil {
		jlog.Errorf("%s, cmd: %d", err, cmd)
		return 0, nil
	}
	raw := encodeFromPack(&Pack{cmd: cmd, data: data})
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(raw))
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return 0, nil
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return 0, nil
	}
	if len(body) == 0 {
		jlog.Errorf("body is empty")
		return 0, nil
	}
	pack := decodeToPack(body)
	if err = proto.Unmarshal(pack.data, out); err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return 0, nil
	}
	return pack.cmd, out
}

func (rpc *Rpc) Response(w http.ResponseWriter, cmd jpb.CMD, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		jlog.Error(err)
		return
	}
	raw := encodeFromPack(&Pack{cmd: cmd, data: data})
	if _, err = w.Write(raw); err != nil {
		jlog.Error(err)
	}
}

// ------------------------- inside -------------------------

func (rpc *Rpc) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := decodeToPack(body)
	han := rpc.handler[pack.cmd]
	if han == nil {
		jlog.Warn("cmd not exist, ", pack.cmd)
		return
	}
	msg := proto.Clone(han.msg)
	if err = proto.Unmarshal(pack.data, msg); err != nil {
		jlog.Warnf("%s, %d", err, pack.cmd)
		return
	}
	han.fun(w, pack.cmd, msg)
}
