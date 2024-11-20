package jnet

import (
	jconfig "jamger/config"
	jhttp "jamger/net/http"
	jkcp "jamger/net/kcp"
	jtcp "jamger/net/tcp"
	jweb "jamger/net/web"
)

// ------------------------- outside -------------------------

var Tcp *jtcp.Tcp
var Kcp *jkcp.Kcp
var Web *jweb.Web
var Http *jhttp.Http

func Run() {
	if cfg := jconfig.Get("tcp"); cfg != nil {
		Tcp = jtcp.NewTcp()
		Tcp.Run()
	}

	if cfg := jconfig.Get("kcp"); cfg != nil {
		Kcp = jkcp.NewKcp()
		Kcp.Run()
	}

	if cfg := jconfig.Get("web"); cfg != nil {
		Web = jweb.NewWeb()
		Web.Run()
	}

	if cfg := jconfig.Get("http"); cfg != nil {
		Http = jhttp.NewHttp()
		Http.Run()
	}
}
