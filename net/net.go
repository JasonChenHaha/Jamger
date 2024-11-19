package jnet

import (
	jconfig "jamger/config"
	jhttp "jamger/net/http"
	jtcp "jamger/net/tcp"
)

// ------------------------- outside -------------------------

var Tcp *jtcp.Tcp
var Http *jhttp.Http

func Run() {
	cfg := jconfig.Get("tcp").(map[string]any)
	if cfg != nil {
		Tcp = jtcp.NewTcp()
		Tcp.Run()
	}

	cfg = jconfig.Get("http").(map[string]any)
	if cfg != nil {
		Http = jhttp.NewHttp()
		Http.Run()
	}
}
