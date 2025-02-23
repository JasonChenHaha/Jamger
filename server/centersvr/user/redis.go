package juser

import (
	"jdb"
	"jglobal"
	"jlog"
)

type Redis struct {
	user *User
	Gate int
}

// ------------------------- package -------------------------

func newRedis(user *User) *Redis {
	re := &Redis{user: user}
	re.load()
	return re
}

func (redis *Redis) load() {
	rsp, err := jdb.Redis.HGet(jglobal.Itoa(redis.user.Id), "gate")
	if err != nil {
		jlog.Error(err)
		return
	}
	if rsp != "" {
		redis.Gate = jglobal.Atoi[int](rsp)
	}
}

// ------------------------- inside -------------------------

func (redis *Redis) SetGate(gate int) {
	redis.Gate = gate
}
