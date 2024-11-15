package main

import (
	jlog "jamger/log"
	jnet "jamger/net"
)

func main() {
	jlog.Info("welcome to jamger!")
	jnet.Run()
}
