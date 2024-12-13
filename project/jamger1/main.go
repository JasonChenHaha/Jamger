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
)

func main() {
	jlog.Info(">jamger start<")
	jetcd.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jwork.Init()
	if jconfig.GetBool("debug") {
		jexample.Init()
	}
	jglobal.Keep()
	jlog.Info(">jamger stop<")
}
