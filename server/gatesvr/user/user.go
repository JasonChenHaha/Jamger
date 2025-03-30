package juser2

import (
	"fmt"
	"jglobal"
	"jschedule"
	"juser"
	"time"
)

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juser.User
	*Redis
	*Mongo
	ticker any
}

var users = jglobal.NewMaps(uint32(1))

// ------------------------- outside -------------------------

func Init() {}

func NewUser(uid uint32) *User {
	user := &User{User: juser.NewUser(uid)}
	user.Redis = newRedis(user)
	user.Mongo = newMongo(user)
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
	return fmt.Sprintf("user(uid=%d,id=%s)", user.Uid, user.Id)
}

func (user *User) Load() *User {
	user.Redis.Load()
	user.Mongo.Load()
	return user
}

func (user *User) Destory() {
	jschedule.Stop(user.ticker)
	users.Delete(user.Uid)
	user.Redis.clear()
	user.User.Flush()
}

// ------------------------- inside -------------------------

func (user *User) tick(args ...any) {
	if user.User.Tick() {
		user.Destory()
	}
}
