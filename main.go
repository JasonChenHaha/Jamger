package main

import (
	jdb "jamger/db"
	jlog "jamger/log"
	jmeta "jamger/meta"
	jnet "jamger/net"
	"jamger/work"
	"os"
	"os/signal"
)

func main() {
	jlog.Info(">jamger start<")

	jdb.Run()

	jnet.Run()

	jmeta.Init()

	work.Run()

	keep()
}

func keep() {
	mainC := make(chan os.Signal, 1)
	signal.Notify(mainC, os.Interrupt)
	<-mainC
	jlog.Info(">jamger shutdown<")
}
