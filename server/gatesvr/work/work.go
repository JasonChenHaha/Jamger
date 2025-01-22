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
		juser.GetUser(rsp.Uid).AesKey = pack.AesKey
		// 缓存
		// if _, err = jdb.Redis.Do("HSET", out["_id"], "token", token); err != nil {
		// 	jlog.Error(err)
		// 	rsp.Code = jpb.CODE_SVR_ERR
		// 	return
		// }
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
