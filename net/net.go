package jnet

import (
	"jconfig"
	"jglobal"
	"jhttp"
	"jkcp"
	"jlog"
	"jnrpc"
	"jpb"
	"jrpc"
	"jtcp"
	"jweb"

	"google.golang.org/protobuf/proto"
)

var Tcp *jtcp.Tcp
var Kcp *jkcp.Kcp
var Web *jweb.Web
var Http *jhttp.Http
var Rpc *jnrpc.Rpc

// ------------------------- outside -------------------------

func Init() {
	Tcp = jtcp.NewTcp()
	if jconfig.Get("tcp") != nil {
		Tcp.AsServer()
	}
	Kcp = jkcp.NewKcp()
	if jconfig.Get("kcp") != nil {
		Kcp.AsServer()
	}
	Web = jweb.NewWeb()
	if jconfig.Get("web") != nil {
		Web.AsServer()
	}
	Http = jhttp.NewHttp()
	if jconfig.Get("http") != nil {
		Http.AsServer()
	}
	Rpc = jnrpc.NewRpc()
	if jconfig.Get("rpc") != nil {
		Rpc.AsServer()
	}
}

// 广播给所有客户端
func BroadcastToC(pack *jglobal.Pack) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	targets := jrpc.Rpc.GetAllTarget(jglobal.GRP_GATE)
	for _, v := range targets {
		v.Proxy(jpb.CMD_BROADCAST, pack)
	}
	return true
}
