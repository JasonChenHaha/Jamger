package main

import (
	"jamger/global"
	jlog "jamger/log"
	jnet "jamger/net"
	"jamger/work"
	"os"
	"os/signal"
)

func main() {
	jlog.Info("jamger start")

	global.G_net = jnet.NewNet()

	work.Init()

	global.G_net.Run()

	keep()
}

func keep() {
	mainC := make(chan os.Signal, 1)
	signal.Notify(mainC, os.Interrupt)
	<-mainC
	jlog.Info("jamger shutdown")
}
