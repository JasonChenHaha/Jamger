package jnet

import (
	"jconfig"
	"jglobal"
	"jhttp"
	"jkcp"
	"jlog"
	"jnrpc"
	"jpb"
	"jrpc"
	"jtcp"
	"jweb"

	"google.golang.org/protobuf/proto"
)

var Tcp *jtcp.Tcp
var Kcp *jkcp.Kcp
var Web *jweb.Web
var Http *jhttp.Http
var Rpc *jnrpc.Rpc
var GetUser func(uint32) any

// ------------------------- outside -------------------------

func Init() {
	Tcp = jtcp.NewTcp()
	if jconfig.Get("tcp") != nil {
		Tcp.AsServer()
	}
	Kcp = jkcp.NewKcp()
	if jconfig.Get("kcp") != nil {
		Kcp.AsServer()
	}
	Web = jweb.NewWeb()
	if jconfig.Get("web") != nil {
		Web.AsServer()
	}
	Http = jhttp.NewHttp()
	if jconfig.Get("http") != nil {
		Http.AsServer()
	}
	Rpc = jnrpc.NewRpc()
	if jconfig.Get("rpc") != nil {
		Rpc.AsServer()
	}
}

func SetGetUser(fun func(uint32) any) {
	GetUser = fun
}

// 广播给所有客户端(->gate->client)
func BroadcastToC(pack *jglobal.Pack) {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return
	}
	targets := jrpc.Rpc.GetAllTarget(jglobal.GRP_GATE)
	data := pack.Data
	for _, v := range targets {
		pack.Data = data
		v.Proxy(jpb.CMD_BROADCAST, pack)
	}
}

// 发给指定客户端(->gate->client)，uids为额外发送列表
func SendToC(pack *jglobal.Pack, uids ...uint32) {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
		return
	}
	data0 := pack.Data
	gate0 := pack.Ctx.(jglobal.User2).GetGate()
	target0 := jrpc.Rpc.GetDirectTarget(jglobal.GRP_GATE, gate0)
	if target0 == nil {
		jlog.Errorf("not find target, gate(%d)", gate0)
		return
	}
	target0.Proxy(jpb.CMD_TOC, pack)
	// 额外发送
	for _, uid := range uids {
		user := GetUser(uid)
		var target *jnrpc.Rpc
		if user == nil {
			target = target0
		} else {
			gate := user.(jglobal.User2).GetGate()
			if gate == gate0 {
				target = target0
			} else {
				target = jrpc.Rpc.GetDirectTarget(jglobal.GRP_GATE, gate)
				if target == nil {
					jlog.Errorf("not find target, gate(%d)", gate)
					continue
				}
			}
		}
		pack.Data = data0
		pack.Ctx = uid
		target.Proxy(jpb.CMD_TOC, pack)
	}
}

// 发送客户端(gate->client)
func Send(pack *jglobal.Pack) {
	user := pack.Ctx.(jglobal.User1)
	tp, id := user.GetSesId()
	switch tp {
	case jglobal.HTTP:
	case jglobal.TCP:
		Tcp.Send(id, pack)
	case jglobal.WEB:
		Web.Send(id, pack)
	case jglobal.KCP:
	}
}

// 关闭连接
func Close(user jglobal.User1) {
	tp, id := user.GetSesId()
	switch tp {
	case jglobal.HTTP:
	case jglobal.TCP:
		Tcp.Close(id)
	case jglobal.WEB:
		Web.Close(id)
	case jglobal.KCP:
	}
}
