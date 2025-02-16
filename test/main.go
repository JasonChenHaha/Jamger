package main

import (
	"jconfig"
	"jglobal"
	"jlog"
)

const (
	packSize     = 2
	uidSize      = 4
	cmdSize      = 2
	checksumSize = 4
	aesKeySize   = 16
)

var aesKey []byte
var uid uint32

func main() {
	jconfig.Init()
	jglobal.Init()
	jlog.Init("")
	var err error
	aesKey, err = jglobal.AesGenerate(16)
	if err != nil {
		jlog.Fatal(err)
	}
	testHttp()
	testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
}
