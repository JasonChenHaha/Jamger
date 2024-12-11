package main

import (
	jwork "jamger1work"
	"jconfig"
	"jdb"
	"jexample"
	"jglobal"
	"jlog"
	"jmeta"
	"jnet"
)

func main() {
	jlog.Info(">jamger start<")
	jdb.Run()

	jnet.Run()

	jmeta.Init()

	jwork.Init()

	if jconfig.GetBool("debug") {
		jexample.Run()
	}

	jglobal.Keep()
	jlog.Info(">jamger stop<")
}
