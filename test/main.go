package main

import (
	"jconfig"
	"jglobal"
	"jlog"
)

const (
	HeadSize     = 2
	CmdSize      = 2
	ChecksumSize = 4
	AesKeySize   = 16
)

func main() {
	jconfig.Init()
	jglobal.Init()
	jlog.Init("")
	// testHttp()
	// testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
}
