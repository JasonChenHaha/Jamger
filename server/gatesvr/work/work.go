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
	jnet.Http.Encoder(httpEncode)
	jnet.Http.Decoder(httpDecode)
	jnet.Rpc.Encoder(rpcEncode)
	jnet.Rpc.Decoder(rpcDecode)
	jnet.Tcp.Encoder(tcpEncode)
	jnet.Tcp.Decoder(tcpDecode)
	jnet.Http.Register(jpb.CMD_PROXY, proxy, nil)
	jnet.Http.Register(jpb.CMD_PING, ping, &jpb.Ping{})
	jnet.Http.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
	jnet.Tcp.Register(jpb.CMD_PROXY, proxy, nil)
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
