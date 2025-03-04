package juser

import (
	"fmt"
	"jglobal"
	"jschedule"
	"juBase"
	"time"
)

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juBase.Base
	*Redis
	*Basic
	ticker any
}

var users = jglobal.NewMaps(uint32(1))

// ------------------------- outside -------------------------

func Init() {}

func NewUser(uid uint32) *User {
	user := &User{Base: juBase.NewBase(uid)}
	user.Redis = newRedis(user)
	user.Basic = newBasic(user)
	user.ticker = jschedule.DoEvery(time.Second, user.tick)
	users.Store(uid, user)
	return user
}

func GetUser(uid uint32) *User {
	if v, ok := users.Load(uid); ok {
		user := v.(*User)
		user.Touch()
		return user
	}
	return nil
}

func Range(fun func(k, v any) bool) {
	users.Range(fun)
}

func (user *User) String() string {
	if user.Basic.Id != "" {
		return fmt.Sprintf("user(%s)", user.Id)
	} else {
		return fmt.Sprintf("user(%d)", user.Uid)
	}
}

func (user *User) Load() *User {
	user.Redis.Load()
	user.Basic.Load()
	return user
}

func (user *User) Destory() {
	jschedule.Stop(user.ticker)
	users.Delete(user.Uid)
	user.Redis.clear()
	user.Base.Flush()
}

// ------------------------- inside -------------------------

func (user *User) tick(args ...any) {
	if user.Base.Tick() {
		user.Destory()
	}
}
