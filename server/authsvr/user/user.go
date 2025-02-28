package juser

import (
	"jschedule"
	"juBase"
	"sync"
	"time"
)

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juBase.Base
	*Redis
	*Basic
	Uid    uint32
	ticker any
}

var users sync.Map

// ------------------------- outside -------------------------

func Init() {}

func GetUser(uid uint32) *User {
	if v, ok := users.Load(uid); ok {
		user := v.(*User)
		user.Touch()
		return user
	} else {
		user := &User{
			Uid:  uid,
			Base: juBase.NewBase(uid),
		}
		user.Basic = newBasic(user)
		user.Redis = newRedis(user)
		user.ticker = jschedule.DoEvery(time.Second, user.tick)
		users.Store(uid, user)
		return user
	}
}

func (user *User) Load() {
	user.Basic.load()
	user.Redis.load()
}

// ------------------------- inside -------------------------

func (user *User) destory() {
	jschedule.Stop(user.ticker)
	users.Delete(user.Uid)
}

func (user *User) tick(args ...any) {
	if user.Base.Tick() {
		user.destory()
	}
}
