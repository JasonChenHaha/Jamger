package jglobal

import (
	"crypto/rsa"
	"fmt"
	"jconfig"
	"jlog"
	"jtrash"
)

var ZONE string
var GROUP string
var SERVER string
var RSA_PRIVATE_KEY *rsa.PrivateKey

const (
	SVR_BEGIN = "nil"
    SVR_BETA = "beta"
    SVR_ALPHA = "alpha"
    SVR_LOGINSVR = "loginsvr"
    SVR_END = "nil"
)

const (
	CMD_OK        = 0
	CMD_ERR       = 1
	CMD_CLOSE     = 2
	CMD_HEARTBEAT = 3
	CMD_PING      = 4
	CMD_PONG      = 5

	CMD_SIGN_UP_REQ = 100
	CMD_SIGN_UP_RSP = 101
	CMD_SIGN_IN_REQ = 102
	CMD_SIGN_IN_RSP = 103
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
	key, err := jtrash.RSALoadPrivateKey(jconfig.GetString("rsa.privateKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	RSA_PRIVATE_KEY = key
}
