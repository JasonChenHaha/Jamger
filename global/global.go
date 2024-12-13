package jglobal

import (
	"os"
	"os/signal"
)

var SVR_NAME = os.Args[0][2:]

const (
	CMD_OK        = 0
	CMD_CLOSE     = 1
	CMD_HEARTBEAT = 2
	CMD_PING      = 3
	CMD_PONG      = 4
)

const (
	EVENT_TEST_1 = 0
	EVENT_TEST_2 = "jamger"
)

type AllInt interface {
	~int | ~uint | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64
}

type AllSInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type AllUInt interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type AllFloat interface {
	~float32 | ~float64
}

func Keep() {
	mainC := make(chan os.Signal, 1)
	signal.Notify(mainC, os.Interrupt)
	<-mainC
}
