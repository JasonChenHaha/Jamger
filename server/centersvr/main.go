package main

import (
	"jconfig"
	"jdb"
	"jetcd"
	"jevent"
	"jglobal"
	"jlog"
	"jmedia"
	"jmeta"
	"jnet"
	"jnet2"
	"jrpc"
	"jschedule"
	"juser"
	"juser2"
	"jwork"
	"jaddress"
)

func main() {
	defer jglobal.Rcover()
	jconfig.Init()
	jglobal.Init()
	jlog.Init(jglobal.SERVER)
	jlog.Infof(">%s start<", jglobal.SERVER)
	jevent.Init()
	jschedule.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jetcd.Init()
	jrpc.Init()
	juser.Init(jrpc.Rpc)
	jnet2.Init()
	juser2.Init()
	jaddress.Init()
	jmedia.Init()
	jwork.Init()
	jglobal.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
