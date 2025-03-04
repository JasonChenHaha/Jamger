package jnrpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
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

type Handler struct {
	fun func(*jglobal.Pack)
	msg proto.Message
}

type Rpc struct {
	addr    string
	client  *http.Client
	handler map[jpb.CMD]*Handler
}

var encoder func(*jglobal.Pack) error
var decoder func(*jglobal.Pack) error

// ------------------------- outside -------------------------

func NewRpc() *Rpc {
	return &Rpc{}
}

func (o *Rpc) String() string {
	return o.addr
}

func (o *Rpc) AsServer() *Rpc {
	o.handler = map[jpb.CMD]*Handler{}
	go func() {
		h2s := &http2.Server{}
		mux := http.NewServeMux()
		mux.HandleFunc("/", o.receive)
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
	return o
}

func (o *Rpc) AsClient(addr string) *Rpc {
	o.addr = "http://" + addr
	o.client = &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
		Timeout: time.Duration(jconfig.GetInt("rpc.timeout")) * time.Millisecond,
	}
	return o
}

func (o *Rpc) Encoder(fun func(*jglobal.Pack) error) {
	encoder = fun
}

func (o *Rpc) Decoder(fun func(*jglobal.Pack) error) {
	decoder = fun
}

func (o *Rpc) Register(cmd jpb.CMD, fun func(pack *jglobal.Pack), msg proto.Message) {
	o.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

// 转发模式
// gate将客户端请求转发至其他group
func (o *Rpc) Transfer(pack *jglobal.Pack) bool {
	if err := encoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	rsp, err := o.client.Post(o.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	if len(body) == 0 {
		jlog.Errorf("body is empty")
		return false
	}
	pack.Data = body
	if err = decoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	return true
}

// 代理模式
// gate将其他group请求发送至客户端
func (o *Rpc) Proxy(cmd jpb.CMD, pack *jglobal.Pack) bool {
	if err := encoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	rsp, err := o.client.Post(fmt.Sprintf("%s/%d", o.addr, cmd), "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	rsp.Body.Close()
	return true
}

// 常规模式
func (o *Rpc) Call(pack *jglobal.Pack, template proto.Message) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	if err = encoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	rsp, err := o.client.Post(o.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	if len(body) == 0 {
		jlog.Errorf("body is empty")
		return false
	}
	pack.Data = body
	if err := decoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	if err = proto.Unmarshal(pack.Data.([]byte), template); err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	pack.Data = template
	return true
}

// ------------------------- inside -------------------------

func (o *Rpc) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{Data: body}
	if err = decoder(pack); err != nil {
		jlog.Error(err)
		return
	}
	cmd := jglobal.UrlToCmd(r.URL.Path)
	switch cmd {
	case jpb.CMD_TOC, jpb.CMD_BROADCAST:
		w.WriteHeader(http.StatusOK)
		han := o.handler[cmd]
		if han == nil {
			jlog.Warnf("cmd(%s) not exist", cmd)
			return
		}
		han.fun(pack)
	default:
		han := o.handler[pack.Cmd]
		if han == nil {
			jlog.Warnf("cmd(%s) not exist", pack.Cmd)
			return
		}
		msg := proto.Clone(han.msg)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.fun(pack)
		if v, ok := pack.Data.(proto.Message); ok {
			pack.Data, err = proto.Marshal(v)
			if err != nil {
				jlog.Error(err)
				return
			}
		}
		if err = encoder(pack); err != nil {
			jlog.Error(err)
			return
		}
		if _, err = w.Write(pack.Data.([]byte)); err != nil {
			jlog.Error(err)
		}
	}
}
