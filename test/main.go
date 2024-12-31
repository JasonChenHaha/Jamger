package main

import (
	"jconfig"
	"jglobal"
	"jlog"
)

const (
	gHeadSize = 2
	gCmdSize  = 2
)

type A struct {
}

func main() {
	jconfig.Init()
	jglobal.Init()
	jlog.Init()
	testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
}
