package jwork

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
	jnet.Tcp.Send(id, 1, []byte("ok!"))
}

func cb2(id uint64, pack *jkcp.Pack) {
	jlog.Debug(jdebug.StructToString(pack))
	jnet.Kcp.Send(id, 1, []byte("ok!"))
}

func cb3(id uint64, pack *jweb.Pack) {
	jlog.Debug(jdebug.StructToString(pack))
	jnet.Kcp.Send(id, 1, []byte("ok!"))
}
