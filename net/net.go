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
	if jconfig.Get("tcp") != nil {
		Tcp = jtcp.NewTcp().AsServer()
	}
	if jconfig.Get("kcp") != nil {
		Kcp = jkcp.NewKcp().AsServer()
	}
	if jconfig.Get("web") != nil {
		Web = jweb.NewWeb()
	}
	if jconfig.Get("http") != nil {
		Http = jhttp.NewHttp().AsServer()
	}
	if jconfig.Get("rpc") != nil {
		Rpc = jnrpc.NewRpc().AsServer()
	}
}
