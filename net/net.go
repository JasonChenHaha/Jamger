package jnet

import (
	"jconfig"
	"jhttp"
	"jkcp"
	"jnrpc"
	"jtcp"
	"jweb"
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
