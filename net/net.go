package jnet

import (
	"jconfig"
	"jhttp"
	"jkcp"
	"jtcp"
	"jweb"
)

var Tcp *jtcp.Tcp
var Kcp *jkcp.Kcp
var Web *jweb.Web
var Http *jhttp.Http

// ------------------------- inside -------------------------

func Init() {
	if jconfig.Get("tcp") != nil {
		Tcp = jtcp.NewTcp()
	}

	if jconfig.Get("kcp") != nil {
		Kcp = jkcp.NewKcp()
	}

	if jconfig.Get("web") != nil {
		Web = jweb.NewWeb()
	}

	if jconfig.Get("http") != nil {
		Http = jhttp.NewHttp()
	}
}
