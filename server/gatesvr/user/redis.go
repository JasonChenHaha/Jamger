package juser

import (
	"bytes"
	"jdb"
	"jglobal"
)

type Redis struct {
	user   *User
	Gate   int
	AesKey []byte
}

// ------------------------- package -------------------------

func newRedis(user *User) *Redis {
	return &Redis{user: user}
}

func (redis *Redis) clear() {
	redis.user.DirtyRedis = nil
}

// ------------------------- outSide -------------------------

func (redis *Redis) Load() *User {
	data, err := jdb.Redis.HGetAll(jglobal.Itoa(redis.user.Uid))
	if err != nil {
		return nil
	}
	if v, ok := data["gate"]; ok {
		redis.Gate = jglobal.Atoi[int](v)
	}
	if v, ok := data["aesKey"]; ok {
		redis.AesKey = []byte(v)
	}
	return redis.user
}

func (redis *Redis) SetGate(gate int) {
	if redis.Gate != gate {
		redis.Gate = gate
		redis.user.Lock()
		redis.user.DirtyRedis["gate"] = gate
		redis.user.UnLock()
	}
}

func (redis *Redis) SetAesKey(aesKey []byte) {
	if !bytes.Equal(redis.AesKey, aesKey) {
		redis.AesKey = aesKey
		redis.user.Lock()
		redis.user.DirtyRedis["aesKey"] = aesKey
		redis.user.UnLock()
	}
}
