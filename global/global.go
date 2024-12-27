package jglobal

import (
	"fmt"
	"jconfig"
)

var ZONE string
var GROUP string
var SERVER string

const (
	SVR_BEGIN = "nil"
    SVR_BETA = "beta"
    SVR_ALPHA = "alpha"
    SVR_END = "nil"
)

const (
	CMD_OK        = 0
	CMD_CLOSE     = 1
	CMD_HEARTBEAT = 2
	CMD_PING      = 3
	CMD_PONG      = 4
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

type AllIntString interface {
	AllInt | string
}

// ------------------------- inside -------------------------

func Init() {
	ZONE = jconfig.GetString("zone")
	GROUP = jconfig.GetString("group")
	SERVER = fmt.Sprintf("%s-%s", GROUP, jconfig.GetString("index"))
}
