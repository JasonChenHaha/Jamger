package main

import (
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
	jglobal.Init()
	jlog.Infof(">%s start<", jglobal.SERVER)
	jetcd.Init()
	jdb.Init()
	jnet.Init()
	jmeta.Init()
	jexample.Init()
	jtrash.Keep()
	jlog.Infof(">%s stop<", jglobal.SERVER)
}
