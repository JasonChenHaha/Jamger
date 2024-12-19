package main

import (
	jwork "gamework"
	"jconfig"
	"jdb"
	"jetcd"
	"jexample"
	"jglobal"
	"jlog"
	"jmeta"
	"jnet"
	"jtrash"
)

func main() {
	jlog.Info(">game start<")
	jglobal.Init()
	jetcd.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jwork.Init()
	if jconfig.GetBool("debug") {
		jexample.Init()
	}
	jtrash.Keep()
	jlog.Info(">game stop<")
}
