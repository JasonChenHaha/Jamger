package work

import (
	jdebug "jamger/debug"
	jlog "jamger/log"
	jnet "jamger/net"
	jkcp "jamger/net/kcp"
	jtcp "jamger/net/tcp"
	jweb "jamger/net/web"
)

func Run() {
	jnet.Tcp.RegisterHandler(1, cb1)
	jnet.Kcp.RegisterHandler(2, cb2)
	jnet.Web.RegisterHandler(1, cb3)
}

func cb1(id uint64, pack *jtcp.Pack) {
	jlog.Debug(jdebug.StructToString(pack))

	pack = &jtcp.Pack{
		Cmd:  1,
		Data: []byte("ok!"),
	}
	jnet.Tcp.Send(id, pack)
}

func cb2(id uint64, pack *jkcp.Pack) {
	jlog.Debug(jdebug.StructToString(pack))

	pack = &jkcp.Pack{
		Cmd:  1,
		Data: []byte("ok!"),
	}
	jnet.Kcp.Send(id, pack)
}

func cb3(id uint64, pack *jweb.Pack) {
	jlog.Debug(jdebug.StructToString(pack))
	jlog.Debug(string(pack.Data))
}
