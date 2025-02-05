package juser

import (
	"jconfig"
	"jlog"
	"sync"
	"time"

	"jschedule"
)

var userMgr *UserMgr

type UserMgr struct {
	user sync.Map
}

// ------------------------- outside -------------------------

func Init() {
	userMgr = &UserMgr{}
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, watch)
	}
}

func GetUser(uid uint32) *User {
	var user *User
	if v, ok := userMgr.user.Load(uid); ok {
		user = v.(*User)
		user.refresh()
	} else {
		user = newUser(uid)
		userMgr.user.Store(uid, user)
	}
	return user
}

// ------------------------- package -------------------------

func delete(uid uint32) {
	userMgr.user.Delete(uid)
}

// ------------------------- debug -------------------------

func watch() {
	counter := 0
	userMgr.user.Range(func(key, value any) bool {
		counter++
		return true
	})
	jlog.Debug("user size: ", counter)
}
