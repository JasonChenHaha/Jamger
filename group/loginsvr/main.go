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
	"jtrash"
	jwork "work"
)

func main() {
	defer jtrash.Rcover()
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
	jwork.Init()
	jtrash.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
