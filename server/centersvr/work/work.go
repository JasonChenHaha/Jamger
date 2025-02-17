package jwork

import (
	"jglobal"
	"jnet"
	"jpb"
	"jrpc"
	"juser"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.Encoder(rpcEncode)
	jnet.Rpc.Decoder(rpcDecode)
	jnet.Rpc.Register(jpb.CMD_PING, ping, &jpb.Ping{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
}

// ------------------------- inside -------------------------

func ping(pack *jglobal.Pack) {
	pack.Cmd = jpb.CMD_PONG
	pack.Data = &jpb.Pong{}
}

// 登录
func login(pack *jglobal.Pack) {
	user := pack.User.(*juser.User)
	rsp := &jpb.LoginRsp{}
	pack.Cmd = jpb.CMD_LOGIN_RSP
	pack.Data = rsp
	user.SetLoginTs()
}
