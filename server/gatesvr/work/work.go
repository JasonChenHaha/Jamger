package jwork

import (
	"jdb"
	"jglobal"
	"jlog"
	"jnet"
	"jpb"
	"jrpc"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Connect(jglobal.GRP_AUTH)
	jrpc.Connect(jglobal.GRP_CENTER)
	jnet.Http.Register(jpb.CMD_PROXY, httpProxy, nil)
	jnet.Http.Register(jpb.CMD_PING, ping, &jpb.Ping{})
	jnet.Http.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
}

// ------------------------- inside -------------------------

func ping(pack *jglobal.Pack) {
	pack.Cmd = jpb.CMD_PONG
	pack.Data = &jpb.Pong{}
}

// 登录
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
		if _, err := jdb.Redis.HSet(jglobal.Itoa(rsp.Uid), "aesKey", pack.AesKey); err != nil {
			jlog.Error(err)
			pack.Cmd = jpb.CMD_GATE_INFO
			pack.Data = &jpb.Error{Code: jpb.CODE_SVR_ERR}
		}
	}
}

// 透传
func httpProxy(pack *jglobal.Pack) {
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
