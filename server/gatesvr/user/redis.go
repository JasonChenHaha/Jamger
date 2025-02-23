package juser

import (
	"bytes"
	"jdb"
	"jglobal"
	"jlog"
)

type Redis struct {
	user   *User
	Gate   int
	AesKey []byte
}

// ------------------------- package -------------------------

func newRedis(user *User) *Redis {
	re := &Redis{user: user}
	re.load()
	return re
}

func (redis *Redis) load() {
	rData, err := jdb.Redis.HGetAll(jglobal.Itoa(redis.user.Uid))
	if err != nil {
		jlog.Error(err)
		return
	}
	if v, ok := rData["gate"]; ok {
		redis.Gate = jglobal.Atoi[int](v)
	}
	if v, ok := rData["aesKey"]; ok {
		redis.AesKey = []byte(v)
	}
}

// ------------------------- inside -------------------------

func (redis *Redis) SetGate(gate int) {
	if redis.Gate != gate {
		redis.Gate = gate
		redis.user.DirtyRedis["gate"] = gate
	}
}

func (redis *Redis) SetAesKey(aesKey []byte) {
	if !bytes.Equal(redis.AesKey, aesKey) {
		redis.AesKey = aesKey
		redis.user.DirtyRedis["aesKey"] = aesKey
	}
}
