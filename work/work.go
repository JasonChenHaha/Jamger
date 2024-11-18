package work

import (
	jdebug "jamger/debug"
	"jamger/global"
	jtcp "jamger/net/tcp"
)

func Init() {
	global.G_net.Tcp.RegisterHandler(1, cb1)
	global.G_net.Tcp.RegisterHandler(2, cb2)
}

func cb1(id uint64, pack jtcp.Pack) {
	jdebug.PrintPack(pack)

	pack = jtcp.Pack{
		Cmd:  1,
		Data: []byte("ok!"),
	}
	global.G_net.Tcp.Send(id, pack)
}

func cb2(id uint64, pack jtcp.Pack) {
	jdebug.PrintPack(pack)

	pack = jtcp.Pack{
		Cmd:  2,
		Data: []byte("ok!"),
	}
	global.G_net.Tcp.Send(id, pack)
}
