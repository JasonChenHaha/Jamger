package jglobal

import (
	"crypto/rsa"
	"fmt"
	"jconfig"
	"jlog"
	"jpb"
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

type User1 interface {
	Destory()
	GetSesId() (int, uint64)
	SetSesId(int, uint64)
}

type User2 interface {
	GetGate() int
}

type Pack struct {
	Cmd  jpb.CMD
	Data any
	Ctx  any
}

const (
	SVR_GATE   = "gatesvr"
	GRP_GATE   = 1
	SVR_AUTH   = "authsvr"
	GRP_AUTH   = 2
	SVR_CENTER = "centersvr"
	GRP_CENTER = 3
)

const (
	MONGO_USER  = "user"
	MONGO_MEDIA = "media"
)

const (
	USER_LIVE = 300
)

const (
	HTTP = 1
	TCP  = 2
	WEB  = 3
	KCP  = 4
)

const (
	UID_SIZE      = 4
	GATE_SIZE     = 4
	CMD_SIZE      = 2
	AESKEY_SIZE   = 16
	CHECKSUM_SIZE = 4
)

var ID int
var NAME string
var ZONE int
var GROUP int
var INDEX int
var SERVER string
var RSA_PRIVATE_KEY *rsa.PrivateKey

// ------------------------- outside -------------------------

func Init() {
	NAME = jconfig.GetString("name")
	ZONE = jconfig.GetInt("zone")
	GROUP = jconfig.GetInt("group")
	INDEX = jconfig.GetInt("index")
	SERVER = fmt.Sprintf("%s-%d", NAME, INDEX)
	ID = ZONE*100000000 + GROUP*10000 + INDEX
	if key := jconfig.GetString("rsa.privateKey"); key != "" {
		key, err := RsaLoadPrivateKey(key)
		if err != nil {
			jlog.Fatal(err)
		}
		RSA_PRIVATE_KEY = key
	}
}

func ParseServerID(id int) (int, int) {
	group := id % 100000000 / 10000
	index := id % 10000
	return group, index
}
