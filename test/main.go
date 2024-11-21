package main

import (
	jlog "jamger/log"
	"strings"
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
	a := "1m"
	b := strings.TrimSuffix(a, "m")
	jlog.Debug(b)
	jlog.Debug(strings.HasSuffix(a, "m"))
}
