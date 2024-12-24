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
	"jtrash"
)

func main() {
	defer jtrash.Rcover()
	jconfig.Init()
	jglobal.Init()
	jlog.Init()
	jlog.Infof(">%s start<", jglobal.SERVER)
	jevent.Init()
	jschedule.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jetcd.Init()
	jrpc.Init()
	jexample.Init()
	jtrash.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
