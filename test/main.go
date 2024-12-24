package main

import (
	"jlog"
)

const (
	gHeadSize = 2
	gCmdSize  = 2
)

func main() {
	jlog.Init()
	jlog.Info("<test start>")
	// testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
}
