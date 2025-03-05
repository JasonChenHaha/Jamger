package juser

import (
	"fmt"
	"jglobal"
	"jnet"
	"jschedule"
	"juBase"
	"time"
)

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juBase.Base
	*Basic
	ticker any
}

var users = jglobal.NewMaps(uint32(1))

// ------------------------- outside -------------------------

func Init() {
	jnet.SetGetUser(GetUserAny)
}

func NewUser(uid uint32) *User {
	user := &User{Base: juBase.NewBase(uid)}
	user.Basic = newBasic(user)
	user.ticker = jschedule.DoEvery(time.Second, user.tick)
	users.Store(uid, user)
	return user
}

func GetUser(uid uint32) *User {
	if v, ok := users.Load(uid); ok {
		user := v.(*User)
		if juBase.Protect.Touch(uid) {
			// 如果处于protect模式，当前user可能是旧的，需要销毁
			user.Destory()
		} else {
			user.Touch()
			return user
		}
	} else {
		// 不存在user仍要touch通知其他节点销毁user
		juBase.Protect.Touch(uid)
	}
	return nil
}

func GetUserAny(uid uint32) any {
	if user := GetUser(uid); user != nil {
		return user
	}
	return nil
}

func Range(fun func(k, v any) bool) {
	users.Range(fun)
}

func (user *User) String() string {
	return fmt.Sprintf("user(uid=%d,id=%s)", user.Uid, user.Id)
}

func (user *User) Load() *User {
	user.Basic.Load()
	return user
}

func (user *User) Destory() {
	jschedule.Stop(user.ticker)
	users.Delete(user.Uid)
	user.Base.Flush()
}

// ------------------------- inside -------------------------

func (user *User) tick(args ...any) {
	if user.Base.Tick() {
		user.Destory()
	}
}
