package main

import (
	"jconfig"
	"jdb"
	"jetcd"
	"jevent"
	"jexample"
	"jglobal"
	"jlog"
	"jmeta"
	"jnet"
	"jrpc"
	"jschedule"
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
	jexample.Init()
	jglobal.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
