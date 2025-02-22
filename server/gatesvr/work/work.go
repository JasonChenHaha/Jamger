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
	jrpc.Connect(jglobal.GRP_AUTH)
	jrpc.Connect(jglobal.GRP_CENTER)
	jnet.Http.Encoder(httpEncode)
	jnet.Http.Decoder(httpDecode)
	jnet.Http.Register(jpb.CMD_PROXY, proxy, nil)
	jnet.Http.Register(jpb.CMD_PING, ping, &jpb.Ping{})
	jnet.Http.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
	jnet.Rpc.Encoder(rpcEncode)
	jnet.Rpc.Decoder(rpcDecode)
	jnet.Rpc.Register(jpb.CMD_PROXY, send, nil)
	jnet.Tcp.Encoder(tcpEncode)
	jnet.Tcp.Decoder(tcpDecode)
	jnet.Tcp.Register(jpb.CMD_PROXY, proxy, nil)
	jnet.Tcp.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
}

// ------------------------- inside -------------------------

func ping(pack *jglobal.Pack) {
	pack.Cmd = jpb.CMD_PONG
	pack.Data = &jpb.Pong{}
}

// 透传
func proxy(pack *jglobal.Pack) {
	target := jrpc.GetRoundRobinTarget(jglobal.GetGroup(pack.Cmd))
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
	target := jrpc.GetRoundRobinTarget(jglobal.GetGroup(pack.Cmd))
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
		// 缓存aesKey
		if _, err := jdb.Redis.HSet(jglobal.Itoa(rsp.Uid), "aesKey", pack.User); err != nil {
			jlog.Error(err)
			pack.Cmd = jpb.CMD_GATE_INFO
			pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		}
	}
}

// 登录
func login(pack *jglobal.Pack) {
	user := pack.User.(*juser.User)
	defer jnet.Tcp.Send(user.SesId, pack)
	// req := pack.Data.(*jpb.LoginReq)
	target := jrpc.GetConsistentHashTarget(jglobal.GRP_CENTER, user.Uid)
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
		user.SetGate(jglobal.ID)
	}
}

// 发送
func send(pack *jglobal.Pack) {
	jnet.Tcp.Send(pack.User.(*juser.User).SesId, pack)
}
