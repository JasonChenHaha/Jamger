package jwork

import (
	"jglobal"
	"jnet"
	"jpb"
	"jrpc"
	"net/http"

	"google.golang.org/protobuf/proto"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Connect(jglobal.GRP_AUTH)
	jnet.Http.SetProxy(httpProxy)
	jnet.Http.Register(jpb.CMD_PING, ping, &jpb.Ping{})
}

// ------------------------- inside -------------------------

func ping(w http.ResponseWriter, cmd jpb.CMD, data proto.Message) {
	rsp := &jpb.Pong{}
	jnet.Http.Response(w, jpb.CMD_PONG, rsp)
}

// 透传
func httpProxy(w http.ResponseWriter, cmd jpb.CMD, data []byte) {
	target := jrpc.GetRoundRobinTarget(jglobal.GetGroup(cmd))
	if target == nil {
		jnet.Http.Response(w, jpb.CMD_GATE_INFO, &jpb.Error{Code: jpb.CODE_SVR_ERR})
		return
	}
	cmd, data = target.Proxy(cmd, data)
	if data == nil {
		jnet.Http.Response(w, jpb.CMD_GATE_INFO, &jpb.Error{Code: jpb.CODE_SVR_ERR})
		return
	}
	jnet.Http.Response(w, cmd, data)
}
