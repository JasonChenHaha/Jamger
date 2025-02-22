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
	"juBase"
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

func (rpc *Rpc) Encoder(fun func(*jglobal.Pack) error) {
	encoder = fun
}

func (rpc *Rpc) Decoder(fun func(*jglobal.Pack) error) {
	decoder = fun
}

func (rpc *Rpc) Register(cmd jpb.CMD, fun func(pack *jglobal.Pack), msg proto.Message) {
	rpc.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

// 转发模式发送
func (rpc *Rpc) Proxy(pack *jglobal.Pack) bool {
	if err := encoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
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

// 请求响应模式发送
func (rpc *Rpc) Call(pack *jglobal.Pack, msg proto.Message) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
		return false
	}
	if err = encoder(pack); err != nil {
		jlog.Error(err)
		return false
	}
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
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
	if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
		return false
	}
	pack.Data = msg
	return true
}

// 请求模式发送
func (rpc *Rpc) Send(pack *jglobal.Pack) {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
		return
	}
	if err = encoder(pack); err != nil {
		jlog.Error(err)
		return
	}
	// rpc.client.Timeout = time.Nanosecond
	rsp, err := rpc.client.Post(rpc.addr, "", bytes.NewBuffer(pack.Data.([]byte)))
	// rpc.client.Timeout = rpc.timeout
	// if err != nil && err != context.DeadlineExceeded {
	// 	jlog.Errorf("%s, %d", err, pack.Cmd)
	// }
	if err != nil {
		jlog.Errorf("%s, %d", err, pack.Cmd)
	}
	rsp.Body.Close()
}

// ------------------------- inside -------------------------

func (rpc *Rpc) receive(w http.ResponseWriter, r *http.Request) {
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
	han := rpc.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, %d", err, pack.Cmd)
			return
		}
		pack.Data = msg
		if pack.User != nil {
			pack.User.(juBase.Locker).Lock()
			han.fun(pack)
			pack.User.(juBase.Locker).UnLock()
		} else {
			han.fun(pack)
		}
	} else {
		han = rpc.handler[jpb.CMD_PROXY]
		if han == nil {
			jlog.Warn("cmd not exist, ", pack.Cmd)
			return
		}
		// 根据业务实际需要决定是否需要加锁
		han.fun(pack)
	}
	if o, ok := pack.Data.(proto.Message); ok {
		pack.Data, err = proto.Marshal(o)
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
