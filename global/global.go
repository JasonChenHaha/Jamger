package jglobal

import (
	"crypto/rsa"
	"jconfig"
	"jlog"
)

var ZONE string
var GROUP string
var INDEX string
var SERVER string
var RSA_PRIVATE_KEY *rsa.PrivateKey

const (
	MONGO_ACCOUNT = "account"
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
	INDEX = jconfig.GetString("index")
	SERVER = GROUP + "-" + INDEX
	key, err := RSALoadPrivateKey(jconfig.GetString("rsa.privateKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	RSA_PRIVATE_KEY = key
}
