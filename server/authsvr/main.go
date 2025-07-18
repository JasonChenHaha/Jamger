package main

import (
	"jconfig"
	"jdb"
	"jetcd"
	"jevent"
	"jglobal"
	"jglobal2"
	"jlog"
	"jmeta"
	"jnet"
	"jnet2"
	"jrpc"
	"jschedule"
	"juser"
	"juser2"
	"jwork"
)

func main() {
	defer jglobal.Rcover()
	jconfig.Init()
	jglobal.Init()
	jglobal2.Init()
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
	jwork.Init()
	jglobal.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
