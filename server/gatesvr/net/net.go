package jnet2

import (
	"jhttp2"
	"jhttps2"
	"jnet"
)

func Init() {
	jnet.Http.SetCodec(httpEncode, httpDecode)
	jnet.Https.SetCodec(httpsEncode, httpsDecode)
	jnet.Tcp.SetCodec(tcpEncode, tcpDecode)
	jnet.Web.SetCodec(webEncode, webDecode)
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jhttp2.Init()
	jhttps2.Init()
}
