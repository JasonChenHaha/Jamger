package jglobal

import (
	"crypto/rsa"
	"fmt"
	"jconfig"
	"jlog"
	"jpb"
	"net/http"
)

var NAME string
var ZONE int
var GROUP int
var INDEX int
var SERVER string
var RSA_PRIVATE_KEY *rsa.PrivateKey

const (
	SVR_GATE   = "gatesvr"
	GRP_GATE   = 1
	SVR_AUTH   = "authsvr"
	GRP_AUTH   = 2
	SVR_CENTER = "centersvr"
	GRP_CENTER = 3
)
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
type Pack struct {
	Cmd    jpb.CMD
	Data   any
	Id     uint64
	W      http.ResponseWriter
	AesKey []byte
}

// ------------------------- outside -------------------------

func Init() {
	NAME = jconfig.GetString("name")
	ZONE = jconfig.GetInt("zone")
	GROUP = jconfig.GetInt("group")
	INDEX = jconfig.GetInt("index")
	SERVER = fmt.Sprintf("%s-%d", NAME, INDEX)
	key, err := RsaLoadPrivateKey(jconfig.GetString("rsa.privateKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	RSA_PRIVATE_KEY = key
}
