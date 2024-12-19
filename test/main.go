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
	a := make([]int, 3)
	a = append(a, 1)
	jlog.Debug(a)
}
