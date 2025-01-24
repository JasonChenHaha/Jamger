package juser

import (
	"jconfig"
	"jdb"
	"jglobal"
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
	jschedule.DoEvery(10*time.Second, tick)
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, watch)
	}
}

func NewUser(uid uint32) *User {
	user := newUser(uid, nil)
	userMgr.user.Store(uid, user)
	return user
}

func GetUser(uid uint32) *User {
	if v, ok := userMgr.user.Load(uid); ok {
		return v.(*User)
	}
	res, err := jdb.Redis.HGetAll(jglobal.Itoa(uid))
	if err != nil {
		jlog.Error(err)
		return nil
	}
	if len(res) > 0 {
		user := newUser(uid, res)
		userMgr.user.Store(uid, user)
		return user
	}
	return nil
}

// ------------------------- inside -------------------------

func tick() {

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
