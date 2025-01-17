package jwork

import (
	"jglobal"
	"jnet"
	"jpb"
	"jrpc"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Connect(jglobal.GRP_AUTH)
	jnet.Tcp.Register(jpb.CMD_PASS, pass, nil)
}

// ------------------------- inside -------------------------

// 透传
func pass(id uint64, cmd jpb.CMD, msg any) {
	target := jrpc.GetRoundRobinTarget(jglobal.GetGroup(cmd))
	if target == nil {
		jnet.Tcp.Send(id, jpb.CMD_GATE_INFO, &jpb.Error{Code: jpb.CODE_SVR_ERR})
		return
	}
	rsp := target.Send(cmd, msg.([]byte))
	if rsp == nil {
		jnet.Tcp.Send(id, jpb.CMD_GATE_INFO, &jpb.Error{Code: jpb.CODE_SVR_ERR})
		return
	}
	jnet.Tcp.Send(id, rsp.Cmd, rsp.Data)
}
