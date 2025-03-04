package jwork

import (
	"jglobal"
	"jlog"
	"jnet"
	"jpb"
	"jrpc"
	"juser"

	"google.golang.org/protobuf/proto"
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
	user := pack.Ctx.(*juser.User)
	rsp := &jpb.LoginRsp{}
	pack.Cmd = jpb.CMD_LOGIN_RSP
	pack.Data = rsp
	user.SetLoginTs()
	// jnet.BroadcastToC(&jglobal.Pack{
	// 	Cmd:  jpb.CMD_NOTIFY,
	// 	Data: &jpb.Notify{Msg: "hello world"},
	// })
	if user.Uid == 1 {
		SendToC(&jglobal.Pack{
			Cmd:  jpb.CMD_NOTIFY,
			Data: &jpb.Notify{Msg: "hello world"},
			Ctx:  user,
		}, 5)
	} else {
		SendToC(&jglobal.Pack{
			Cmd:  jpb.CMD_NOTIFY,
			Data: &jpb.Notify{Msg: "hello world"},
			Ctx:  user,
		}, 1)
	}
}

// 发给指定客户端
func SendToC(pack *jglobal.Pack, uids ...uint32) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return false
	}
	user0 := pack.Ctx.(*juser.User)
	uid0 := user0.Uid
	for _, uid := range uids {
		user := juser.HasUser(uid)
		if user == nil {
			user = user0
			user.Uid = uid
		}
		pack.Ctx = user
		target := jrpc.Rpc.GetDirectTarget(jglobal.GRP_GATE, user.Gate)
		if target == nil {
			jlog.Warnf("can't find target, group(%d), index(%d)", jglobal.GRP_GATE, user.Gate)
			continue
		}
		target.Proxy(jpb.CMD_TOC, pack)
		user.Uid = uid
	}
	user0.Uid = uid0
	return true
}
