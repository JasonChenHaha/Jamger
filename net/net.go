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
	if jconfig.Get("tcp") != nil {
		Tcp = jtcp.NewTcp()
		Tcp.Run()
	}

	if jconfig.Get("kcp") != nil {
		Kcp = jkcp.NewKcp()
		Kcp.Run()
	}

	if jconfig.Get("web") != nil {
		Web = jweb.NewWeb()
		Web.Run()
	}

	if jconfig.Get("http") != nil {
		Http = jhttp.NewHttp()
		Http.Run()
	}
}
