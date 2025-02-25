package jnet

import (
	"jconfig"
	"jglobal"
	"jhttp"
	"jkcp"
	"jlog"
	"jnrpc"
	"jrpc"
	"jtcp"
	"juser"
	"jweb"

	"google.golang.org/protobuf/proto"
)

var Tcp *jtcp.Tcp
var Kcp *jkcp.Kcp
var Web *jweb.Web
var Http *jhttp.Http
var Rpc *jnrpc.Rpc

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

// 广播给所有客户端
func BroadcastToC(pack *jglobal.Pack) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%d)", err, pack.Cmd)
		return false
	}
	targets := jrpc.GetAllTarget(jglobal.GRP_GATE)
	for _, v := range targets {
		v.BroadcastToC(pack)
	}
	return true
}

// 发给指定客户端
func SendToC(pack *jglobal.Pack, ids ...uint32) bool {
	var err error
	pack.Data, err = proto.Marshal(pack.Data.(proto.Message))
	if err != nil {
		jlog.Errorf("%s, cmd(%d)", err, pack.Cmd)
		return false
	}
	for _, id := range ids {
		// to do: 这里只需要用到user.Gate，不需要加载完整的user
		user := juser.GetUser(id)
		if user.Gate != 0 {
			group, index := jglobal.ParseServerID(user.Gate)
			target := jrpc.GetDirectTarget(group, index)
			if target != nil {
				pack.Ctx = user
				target.SendToC(pack)
			} else {
				jlog.Warnf("can't find target, group(%d), index(%d)", group, index)
			}
		} else {
			jlog.Warnf("can't find user's(%d) gate", id)
		}
	}
	return true
}
