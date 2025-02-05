package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jschedule"
	"time"
)

const (
	EXPIRE = 60 * 10
)

// 所有属性的写需要使用对应的set方法来驱动定时落地
type User struct {
	Uid    uint32
	AesKey []byte
	expire int
	redis  bool
}

// ------------------------- package -------------------------

func newUser(uid uint32) *User {
	user := &User{
		Uid:    uid,
		expire: EXPIRE,
	}
	res, err := jdb.Redis.HGetAll(jglobal.Itoa(uid))
	if err != nil {
		jlog.Error(err)
		return nil
	}
	if len(res) > 0 {
		user.AesKey = []byte(res["aesKey"])
	}
	jschedule.DoEvery(time.Second, user.tick)
	return user
}

func (user *User) refresh() {
	user.expire = EXPIRE
}

// ------------------------- outside -------------------------

func (user *User) SetAesKey(key []byte) {
	user.AesKey = key
	user.redis = true
}

// ------------------------- inside -------------------------

func (user *User) tick() {
	if user.redis {
		if _, err := jdb.Redis.HSet(jglobal.Itoa(user.Uid), "aesKey", user.AesKey); err != nil {
			jlog.Error(err)
		} else {
			user.redis = false
		}
	}
	if user.expire -= 1; user.expire <= 0 {
		delete(user.Uid)
	}
}
