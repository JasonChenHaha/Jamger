package juser

import (
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
	return re
}

func (redis *Redis) Load() {
	if redis.AesKey != nil {
		return
	}
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
	redis.Gate = gate
	redis.user.DirtyRedis["gate"] = gate
}

func (redis *Redis) SetAesKey(aesKey []byte) {
	redis.AesKey = aesKey
	redis.user.DirtyRedis["aesKey"] = aesKey
}
