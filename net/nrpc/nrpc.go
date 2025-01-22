package jnrpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
)

type Func func(pack *jglobal.Pack)

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

func (rpc *Rpc) Proxy(pack *jglobal.Pack) bool {
	encodePack(pack)
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	if len(body) == 0 {
		jlog.Errorf("body is empty")
		return false
	}
	decodeToPack(pack, body)
	return true
}

func (rpc *Rpc) Call(pack *jglobal.Pack, msg proto.Message) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
		return false
	}
	encodePack(pack)
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	if len(body) == 0 {
		jlog.Errorf("body is empty")
		return false
	}
	decodeToPack(pack, body)
	if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	pack.Data = msg
	return true
}

// ------------------------- inside -------------------------

func (rpc *Rpc) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{W: w}
	decodeToPack(pack, body)
	han := rpc.handler[pack.Cmd]
	if han == nil {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		return
	}
	msg := proto.Clone(han.msg)
	if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
		jlog.Warnf("%s, %d", err, pack.Cmd)
		return
	}
	pack.Data = msg
	han.fun(pack)
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Error(err)
		return
	}
	encodePack(pack)
	if _, err = pack.W.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}
