package main

import (
	"jlog"
)

const (
	gHeadSize = 2
	gCmdSize  = 2
)

func main() {
	jlog.Info("<test start>")
	// testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
	k := test().(*ABC)
	jlog.Debug(k)
}

type ABC struct{}

func test() any {
	return nil
}
