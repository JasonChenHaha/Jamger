package main

import (
	"fmt"
	"hash/fnv"
	"jconfig"
	"jglobal"
	"jlog"
	"os"
)

const (
	packSize     = 2
	uidSize      = 4
	cmdSize      = 2
	checksumSize = 4
	aesKeySize   = 16
)

var gateNum = 1
var aesKey []byte
var uid uint32
var id string
var pwd string
var httpAddr string
var tcpAddr string

func main() {
	jconfig.Init()
	jglobal.Init()
	jlog.Init("")
	id, pwd = os.Args[2], os.Args[3]
	httpAddr = makeAddr(jconfig.GetString("http.addr"), id)
	tcpAddr = makeAddr(jconfig.GetString("tcp.addr"), id)
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

func makeAddr(addr string, key string) string {
	h := fnv.New32a()
	h.Write([]byte(key))
	n := h.Sum32() % uint32(gateNum)
	return fmt.Sprintf("%s%d", addr[0:len(addr)-1], n)
}
