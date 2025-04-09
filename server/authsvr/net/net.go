package jnet2

import (
	"jconfig"
	"jnet"
)

func Init() {
	if jconfig.Get("rpc") != nil {
		jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	}
}
