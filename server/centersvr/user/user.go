package juser2

import (
	"fmt"
	"jconfig"
	"jglobal"
	"jlog"
	"jnet"
	"jschedule"
	"juser"
	"time"
)

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juser.User
	*Mongo
	ticker any
}

var users = jglobal.NewMaps(uint32(1))

// ------------------------- outside -------------------------

func Init() {
	jnet.SetGetUser(GetUserAny)
	if jconfig.Get("debug") != nil {
		jschedule.DoEvery(time.Duration(jconfig.GetInt("debug.interval"))*time.Millisecond, func(args ...any) {
			jlog.Debugf("center user %v", users)
		})
	}
}

func NewUser(uid uint32) *User {
	user := &User{User: juser.NewUser(uid)}
	user.Mongo = newMongo(user)
	user.ticker = jschedule.DoEvery(time.Second, user.tick)
	users.Store(uid, user)
	return user
}

func GetUser(uid uint32) *User {
	if v, ok := users.Load(uid); ok {
		user := v.(*User)
		if juser.Protect.Touch(uid) {
			// 如果处于protect模式，当前user可能是旧的，需要销毁
			user.Destory()
		} else {
			user.Touch()
			return user
		}
	} else {
		// 不存在user仍要touch通知其他节点销毁user
		juser.Protect.Touch(uid)
	}
	return nil
}

func GetUserAnyway(uid uint32) *User {
	user := GetUser(uid)
	if user == nil {
		user = NewUser(uid).Load()
	}
	return user
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
	return fmt.Sprintf("user(uid=%d,id=%s)", user.Uid, user.Basic.Id)
}

func (user *User) Load() *User {
	user.Mongo.Load()
	return user
}

func (user *User) Destory() {
	jschedule.Stop(user.ticker)
	users.Delete(user.Uid)
	user.User.Flush()
}

// ------------------------- inside -------------------------

func (user *User) tick(args ...any) {
	if user.User.Tick() {
		user.Destory()
	}
}
