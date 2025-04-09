package juser2

import (
	"bytes"
	"fmt"
	"jdb"
)

type Redis struct {
	user   *User
	Gate   int
	Token  string
	AesKey []byte
}

// ------------------------- package -------------------------

func newRedis(user *User) *Redis {
	return &Redis{user: user}
}

func (redis *Redis) clear() {
	redis.user.Lock()
	redis.user.DirtyRedis = nil
	redis.user.UnLock()
}

// ------------------------- outSide -------------------------

func (redis *Redis) Load() *User {
	// data, err := jdb.Redis.HGetAll(jglobal.Itoa(redis.user.Uid))
	// if err != nil {
	// 	return nil
	// }
	// if v, ok := data["gate"]; ok {
	// 	redis.Gate = jglobal.Atoi[int](v)
	// }
	// if v, ok := data["aesKey"]; ok {
	// 	redis.AesKey = []byte(v)
	// }
	token, err := jdb.Redis.Get(fmt.Sprintf("%d-token", redis.user.Uid))
	if err != nil {
		return nil
	}
	redis.Token = token
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

func (redis *Redis) SetToken(token string) {
	if token != redis.Token {
		redis.Token = token
		redis.user.Lock()
		redis.user.DirtyRedis["token"] = token
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
