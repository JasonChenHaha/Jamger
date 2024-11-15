package main

import (
	jconfig "jamger/config"
	jlog "jamger/log"
)

func main() {
	jlog.Infoln("welcome to jamger!")
	err := jconfig.Load()
	if err != nil {
		jlog.Errorln(err)
	}
}
