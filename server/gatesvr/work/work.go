package jwork

import (
	"jdb"
	"jglobal"
	"jlog"
	"jnet"
	"jpb"
	"jrpc"
	"juser"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jrpc.Rpc.Connect(jglobal.GRP_AUTH)
	jrpc.Rpc.Connect(jglobal.GRP_CENTER)
	jnet.Http.SetCodec(httpEncode, httpDecode)
	jnet.Http.Register(jpb.CMD_TRANSFER, httpTransfer, nil)
	jnet.Http.Register(jpb.CMD_SIGN_IN_REQ, httpSignIn, &jpb.SignInReq{})
	jnet.Http.Register(jpb.CMD_LOGIN_REQ, httpLogin, &jpb.LoginReq{})
	jnet.Https.SetCodec(httpsEncode, httpsDecode)
	jnet.Https.Register(jpb.CMD_TRANSFER, httpTransfer, nil)
	jnet.Https.Register(jpb.CMD_IMAGE_REQ, httpImage, nil)
	jnet.Tcp.SetCodec(tcpEncode, tcpDecode)
	jnet.Tcp.Register(jpb.CMD_HEARTBEAT, twHeartbeat, &jpb.HeartbeatReq{})
	jnet.Tcp.Register(jpb.CMD_TRANSFER, twTransfer, nil)
	jnet.Tcp.Register(jpb.CMD_LOGIN_REQ, twLogin, &jpb.LoginReq{})
	jnet.Web.SetCodec(webEncode, webDecode)
	jnet.Web.Register(jpb.CMD_HEARTBEAT, twHeartbeat, &jpb.HeartbeatReq{})
	jnet.Web.Register(jpb.CMD_TRANSFER, twTransfer, nil)
	jnet.Web.Register(jpb.CMD_LOGIN_REQ, twLogin, &jpb.LoginReq{})
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jnet.Rpc.Register(jpb.CMD_KICK_USER_REQ, rpcKickUser, &jpb.KickUserReq{})
	jnet.Rpc.Register(jpb.CMD_TOC, rpcSendToC, nil)
	jnet.Rpc.Register(jpb.CMD_BROADCAST, rpcBroadcast, nil)
}

// ------------------------- inside.method.http/https -------------------------

// 转发
func httpTransfer(pack *jglobal.Pack) {
	target := jrpc.Rpc.GetRoundRobinTarget(jglobal.GetGroup(pack.Cmd))
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "can't find target"}
		return
	}
	if !target.Transfer(pack) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "transfer failed"}
	}
}

// auth登录
func httpSignIn(pack *jglobal.Pack) {
	target := jrpc.Rpc.GetRoundRobinTarget(jglobal.GetGroup(pack.Cmd))
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "can't find target"}
		return
	}
	if !target.Call(pack, &jpb.SignInRsp{}) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "call failed"}
		return
	}
	rsp := pack.Data.(*jpb.SignInRsp)
	if rsp.Code == jpb.CODE_OK {
		uid := jglobal.Itoa(rsp.Uid)
		gate, err := jdb.Redis.HGet(uid, "gate")
		if err != nil {
			pack.Cmd = jpb.CMD_GATE_INFO
			pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "get gate failed"}
			return
		}
		if gate != "" {
			// 旧连接踢下线
			target = jrpc.Rpc.GetDirectTarget(jglobal.GROUP, jglobal.Atoi[int](gate))
			if target == nil {
				pack.Cmd = jpb.CMD_GATE_INFO
				pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "can't find target"}
				return
			}
			pack2 := &jglobal.Pack{
				Cmd:  jpb.CMD_KICK_USER_REQ,
				Data: &jpb.KickUserReq{Uid: rsp.Uid},
			}
			if !target.Call(pack2, &jpb.KickUserRsp{}) {
				pack.Cmd = jpb.CMD_GATE_INFO
				pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "call failed"}
				return
			}
			if pack2.Data.(*jpb.KickUserRsp).Code != jpb.CODE_OK {
				pack.Cmd = jpb.CMD_GATE_INFO
				pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "kick user failed"}
				return
			}
		}
		// 缓存aesKey
		if _, err := jdb.Redis.HSet(uid, "aesKey", pack.Ctx); err != nil {
			pack.Cmd = jpb.CMD_GATE_INFO
			pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "save aeskey failed"}
		}
	}
}

// 登录
func httpLogin(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	target := jrpc.Rpc.GetConsistentHashTarget(jglobal.GRP_CENTER, user.Uid)
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "can't find target"}
		return
	}
	if !target.Call(pack, &jpb.LoginRsp{}) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "call failed"}
		return
	}
}

// 图片下载
func httpImage(pack *jglobal.Pack) {
	target := jrpc.Rpc.GetRoundRobinTarget(jglobal.GRP_CENTER)
	if target == nil {
		jlog.Error("can't find target")
		return
	}
	if !target.Call(pack, &jpb.ImageRsp{}) {
		jlog.Error("call failed")
		return
	}
	rsp := pack.Data.(*jpb.ImageRsp)
	if rsp.Code != jpb.CODE_OK {
		pack.Data = nil
		jlog.Errorf("error code(%s)", rsp.Code)
		return
	}
	pack.Data = rsp.Image
}

// ------------------------- inside.method.tcp/web -------------------------

// 透传
func twTransfer(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	defer jnet.Send(pack)
	target := jrpc.Rpc.GetConsistentHashTarget(jglobal.GetGroup(pack.Cmd), user.Uid)
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "can't find target"}
		return
	}
	if !target.Transfer(pack) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "transfer failed"}
	}
}

// 心跳
func twHeartbeat(pack *jglobal.Pack) {
	// do nothing
}

// 登录
func twLogin(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	defer jnet.Send(pack)
	user.SetGate(jglobal.INDEX)
	target := jrpc.Rpc.GetConsistentHashTarget(jglobal.GRP_CENTER, user.Uid)
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "can't find target"}
		return
	}
	if !target.Call(pack, &jpb.LoginRsp{}) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR, Desc: "call failed"}
		return
	}
}

// ------------------------- inside.method.rpc -------------------------

// 踢下线
func rpcKickUser(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.KickUserReq)
	pack.Cmd = jpb.CMD_KICK_USER_RSP
	pack.Data = &jpb.KickUserRsp{}
	if user := juser.GetUser(req.Uid); user != nil {
		jnet.Close(user)
		user.Destory()
	}
}

// 发送
func rpcSendToC(pack *jglobal.Pack) {
	switch v := pack.Ctx.(type) {
	case *juser.User:
		jnet.Send(pack)
	case uint32:
		// 转发
		var user *juser.User
		if user = juser.GetUser(v); user == nil {
			user = juser.NewUser(v).Redis.Load()
		}
		if user.Gate == 0 {
			jlog.Warnf("%s is offline", user)
			return
		}
		target := jrpc.Rpc.GetDirectTarget(jglobal.GRP_GATE, user.Gate)
		if target == nil {
			return
		}
		pack.Ctx = user
		target.Proxy(jpb.CMD_TOC, pack)
	}
}

// 广播
func rpcBroadcast(pack *jglobal.Pack) {
	data := pack.Data
	juser.Range(func(uid, user any) bool {
		pack.Data = data
		pack.Ctx = user
		jnet.Send(pack)
		return true
	})
}
