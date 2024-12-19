package main

import (
	jwork "gatework"
	"jconfig"
	"jdb"
	"jetcd"
	"jexample"
	"jglobal"
	"jlog"
	"jmeta"
	"jnet"
	"jtrash"
	"os"
)

func main() {
	jlog.Info(">gate start<")
	jglobal.Init()
	os.Args[0] = jglobal.SERVER
	jetcd.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jwork.Init()
	if jconfig.GetBool("debug") {
		jexample.Init()
	}
	jtrash.Keep()
	jlog.Info(">gate stop<")
}
