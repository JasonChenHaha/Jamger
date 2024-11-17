package main

import (
	jlog "jamger/log"
	jnet "jamger/net"
	"jamger/net/tcp"
)

func main() {
	jlog.Info("welcome to jamger!")
	jnet.SetCallback(cb)
	jnet.Run()
}

func cb(id uint64, pack tcp.Pack) {

}
