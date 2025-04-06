package jnet2

import (
	"jconfig"
	"jhttp2"
	"jhttps2"
	"jnet"
)

func Init() {
	if jconfig.Get("tcp") != nil {
		jnet.Tcp.SetCodec(tcpEncode, tcpDecode)
	}
	if jconfig.Get("web") != nil {
		jnet.Web.SetCodec(webEncode, webDecode)
	}
	if jconfig.Get("http") != nil {
		jnet.Http.SetCodec(httpEncode, httpDecode)
		jhttp2.Init()
	}
	if jconfig.Get("https") != nil {
		jnet.Https.SetCodec(httpsEncode, httpsDecode)
		jhttps2.Init()
	}
	if jconfig.Get("rpc") != nil {
		jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	}
}
