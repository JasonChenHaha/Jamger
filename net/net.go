package jnet

import (
	jconfig "jamger/config"
	jhttp "jamger/net/http"
	jtcp "jamger/net/tcp"
)

type Net struct {
	Tcp  *jtcp.Tcp
	Http *jhttp.Http
}

// ------------------------- outside -------------------------

func NewNet() *Net {
	net := &Net{}
	cfg := jconfig.Get("tcp").(map[string]any)
	if cfg != nil {
		net.Tcp = jtcp.NewTcp(cfg["addr"].(string))
	}

	cfg = jconfig.Get("http").(map[string]any)
	if cfg != nil {

		net.Http = jhttp.NewHttp(cfg["addr"].(string))
	}

	return net
}

func (net *Net) Run() {
	net.Tcp.Run()
}
