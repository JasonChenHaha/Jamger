package main

import (
	jwork "jamger1work"
	"jconfig"
	"jdb"
	"jetcd"
	"jexample"
	"jglobal"
	"jlog"
	"jmeta"
	"jnet"
	"jrpc"
	"jtrash"
)

func main() {
	jlog.Info(">jamger start<")
	jglobal.Init()
	jetcd.Init()
	jrpc.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jwork.Init()
	if jconfig.GetBool("debug") {
		jexample.Init()
	}
	jtrash.Keep()
	jlog.Info(">jamger stop<")
}
