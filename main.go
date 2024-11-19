package main

import (
	jlog "jamger/log"
	jnet "jamger/net"
	"jamger/work"
	"os"
	"os/signal"
)

func main() {
	jlog.Info(">jamger start<")

	jnet.Run()

	work.Init()

	keep()
}

func keep() {
	mainC := make(chan os.Signal, 1)
	signal.Notify(mainC, os.Interrupt)
	<-mainC
	jlog.Info(">jamger shutdown<")
}
