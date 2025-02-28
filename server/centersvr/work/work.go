package jwork

import (
	"jglobal"
	"jlog"
	"jnet"
	"jpb"
	"jrpc"
	"juser"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_CENTER)
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.Encoder(rpcEncode)
	jnet.Rpc.Decoder(rpcDecode)
	jnet.Rpc.Register(jpb.CMD_DEL_USER, deleteUser, &jpb.DelUserReq{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
}

// ------------------------- inside -------------------------

// 缓存清理
func deleteUser(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.DelUserReq)
	pack.Data = &jpb.DelUserRsp{}
	juser.DelUser(req.Uid)
}

// 登录
func login(pack *jglobal.Pack) {
	jlog.Debug("login!!!")
	user := pack.Ctx.(*juser.User)
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
	// jnet.SendToC(&jglobal.Pack{
	// 	Cmd:  jpb.CMD_NOTIFY,
	// 	Data: &jpb.Notify{Msg: "hello world"},
	// }, 2)
	// }
}
