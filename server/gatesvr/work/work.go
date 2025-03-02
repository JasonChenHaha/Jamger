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
	jnet.Http.Encoder(httpEncode)
	jnet.Http.Decoder(httpDecode)
	jnet.Http.Register(jpb.CMD_PROXY, httpProxy, nil)
	jnet.Http.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
	jnet.Tcp.Encoder(tcpEncode)
	jnet.Tcp.Decoder(tcpDecode)
	jnet.Tcp.Register(jpb.CMD_HEARTBEAT, heartbeat, &jpb.HeartbeatReq{})
	jnet.Tcp.Register(jpb.CMD_PROXY, tcpProxy, nil)
	jnet.Tcp.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
	jnet.Rpc.Encoder(rpcEncode)
	jnet.Rpc.Decoder(rpcDecode)
	jnet.Rpc.Register(jpb.CMD_KICK_USER_REQ, kickUser, &jpb.KickUserReq{})
	jnet.Rpc.Register(jpb.CMD_TOC, sendToC, nil)
	jnet.Rpc.Register(jpb.CMD_BROADCAST, broadcast, nil)
}

// ------------------------- inside.http -------------------------

// 透传
func httpProxy(pack *jglobal.Pack) {
	target := jrpc.Rpc.GetRoundRobinTarget(jglobal.GetGroup(pack.Cmd))
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		return
	}
	if !target.Proxy(pack) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
	}
}

// auth登录
func signIn(pack *jglobal.Pack) {
	target := jrpc.Rpc.GetRoundRobinTarget(jglobal.GetGroup(pack.Cmd))
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		return
	}
	if !target.Call(pack, &jpb.SignInRsp{}) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		return
	}
	rsp := pack.Data.(*jpb.SignInRsp)
	if rsp.Code == jpb.CODE_OK {
		uid := jglobal.Itoa(rsp.Uid)
		gate, err := jdb.Redis.HGet(uid, "gate")
		if err != nil {
			pack.Cmd = jpb.CMD_GATE_INFO
			pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
			return
		}
		if gate != "" {
			// 旧连接踢下线
			target = jrpc.Rpc.GetDirectTarget(jglobal.GROUP, jglobal.Atoi[int](gate))
			if target == nil {
				pack.Cmd = jpb.CMD_GATE_INFO
				pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
				return
			}
			pack2 := &jglobal.Pack{
				Cmd:  jpb.CMD_KICK_USER_REQ,
				Data: &jpb.KickUserReq{Uid: rsp.Uid},
			}
			if !target.Call(pack2, &jpb.KickUserRsp{}) {
				pack.Cmd = jpb.CMD_GATE_INFO
				pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
				return
			}
			if pack2.Data.(*jpb.KickUserRsp).Code != jpb.CODE_OK {
				pack.Cmd = jpb.CMD_GATE_INFO
				pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
				return
			}
		}
		// 缓存aesKey
		if _, err := jdb.Redis.HSet(uid, "aesKey", pack.Ctx); err != nil {
			jlog.Error(err)
			pack.Cmd = jpb.CMD_GATE_INFO
			pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		}
	}
}

// ------------------------- inside.tcp -------------------------

// 透传
func tcpProxy(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	defer jnet.Tcp.Send(pack)
	target := jrpc.Rpc.GetConsistentHashTarget(jglobal.GetGroup(pack.Cmd), user.Uid)
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		return
	}
	if !target.Proxy(pack) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
	}
}

// 心跳
func heartbeat(pack *jglobal.Pack) {
	// do nothing
}

// 登录
func login(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	defer jnet.Tcp.Send(pack)
	target := jrpc.Rpc.GetConsistentHashTarget(jglobal.GRP_CENTER, user.Uid)
	if target == nil {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		return
	}
	if !target.Call(pack, &jpb.LoginRsp{}) {
		pack.Cmd = jpb.CMD_GATE_INFO
		pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		return
	}
	rsp := pack.Data.(*jpb.LoginRsp)
	if rsp.Code == jpb.CODE_OK {
		user.SetGate(jglobal.INDEX)
	}
}

// ------------------------- inside.rpc -------------------------

// 踢下线
func kickUser(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.KickUserReq)
	pack.Cmd = jpb.CMD_KICK_USER_RSP
	pack.Data = &jpb.KickUserRsp{}
	user := juser.GetUser(req.Uid)
	jnet.Tcp.Close(user.SesId)
}

// 发送
func sendToC(pack *jglobal.Pack) {
	jnet.Tcp.Send(pack)
}

// 广播
func broadcast(pack *jglobal.Pack) {
	juser.Range(func(k, v any) bool {
		p := &jglobal.Pack{
			Cmd:  pack.Cmd,
			Data: pack.Data,
			Ctx:  v,
		}
		jnet.Tcp.Send(p)
		return true
	})
}
