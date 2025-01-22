package jwork

import (
	"jglobal"
	"jpb"
)

// ------------------------- outside -------------------------

func Init() {
	// jnet.Rpc.Register(jpb.CMD_PING, ping, &jpb.Ping{})
}

// ------------------------- inside -------------------------

func ping(pack *jglobal.Pack) {
	pack.Cmd = jpb.CMD_PONG
	pack.Data = &jpb.Pong{}
}
