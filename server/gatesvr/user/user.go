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
	SesId  uint64
	ticker any
	destro bool
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

func Range(fun func(k, v any) bool) {
	users.Range(fun)
}

func (user *User) Load() {
	user.Basic.load()
	user.Redis.load()
}

func (user *User) GetSesId() uint64 {
	return user.SesId
}

func (user *User) SetSesId(id uint64) {
	user.SesId = id
}

func (user *User) Destory() {
	user.destro = true
}

// ------------------------- inside -------------------------

func (user *User) tick(args ...any) {
	if user.Base.Tick() {
		user.destro = true
	}
	if user.destro {
		jschedule.Stop(user.ticker)
		users.Delete(user.Uid)
	}
}
