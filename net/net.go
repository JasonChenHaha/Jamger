package jnet

import (
	"jconfig"
	"jhttp"
	"jkcp"
	"jtcp"
	"jweb"
)

var Tcp *jtcp.TcpSvr
var Kcp *jkcp.KcpSvr
var Web *jweb.WebSvr
var Http *jhttp.HttpSvr

// ------------------------- inside -------------------------

func Init() {
	if jconfig.Get("tcp") != nil {
		Tcp = jtcp.NewTcpSvr()
	}
	if jconfig.Get("kcp") != nil {
		Kcp = jkcp.NewKcpSvr()
	}
	if jconfig.Get("web") != nil {
		Web = jweb.NewWebSvr()
	}
	if jconfig.Get("http") != nil {
		Http = jhttp.NewHttpSvr()
	}
}
