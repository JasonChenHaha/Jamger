package main

import (
	jlog "jamger/log"
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

	select {}
}
