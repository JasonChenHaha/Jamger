package main

import (
	"jconfig"
	"jdb"
	"jetcd"
	"jevent"
	"jglobal"
	"jlog"
	"jmeta"
	"jnet"
	"jrpc"
	"jschedule"
	"juser2"
	"jwork"
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
	juser2.Init()
	jwork.Init()
	jglobal.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
