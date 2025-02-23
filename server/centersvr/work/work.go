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
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
}

// ------------------------- inside -------------------------

// 登录
func login(pack *jglobal.Pack) {
	user := pack.User.(*juser.User)
	rsp := &jpb.LoginRsp{}
	pack.Cmd = jpb.CMD_LOGIN_RSP
	pack.Data = rsp
	user.SetLoginTs()
	jnet.BroadcastToC(&jglobal.Pack{
		Cmd:  jpb.CMD_NOTIFY,
		Data: &jpb.Notify{Msg: "hello world"},
	})
	// if user.Uid == 1 {
	// 	jnet.SendToC(&jglobal.Pack{
	// 		Cmd:  jpb.CMD_NOTIFY,
	// 		Data: &jpb.Notify{Msg: "hello world"},
	// 	}, 2)
	// } else {
	// 	jnet.SendToC(&jglobal.Pack{
	// 		Cmd:  jpb.CMD_NOTIFY,
	// 		Data: &jpb.Notify{Msg: "hello world"},
	// 	}, 2)
	// }
}
