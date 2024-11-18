package jdebug

import (
	jlog "jamger/log"
	jtcp "jamger/net/tcp"
)

func PrintPack(pack jtcp.Pack) {
	jlog.Debug("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	jlog.Debug("cmd: ", pack.Cmd)
	jlog.Debug("data: ", string(pack.Data))
	jlog.Debug("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
}
