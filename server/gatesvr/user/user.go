package juser

import (
	"jlog"
	"jschedule"
	"juBase"
	"sync"
	"time"
)

var users sync.Map

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juBase.Base
	*Redis
	*Basic
	Uid    uint32
	SesId  uint64
	ticker any
}

// ------------------------- outside -------------------------

func Init() {}

func GetUser(uid uint32) *User {
	if v, ok := users.Load(uid); ok {
		user := v.(*User)
		user.Refresh()
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

// func (user *User) Send(pack *jglobal.Pack) {
// 	target := jrpc.GetDirectTarget(jglobal.ParseServerID(user.Gate))
// 	if target == nil {
// 		jlog.Errorf("no target, serverID: %d", user.Gate)
// 		return
// 	}
// 	target.Send(pack)
// }

func (user *User) SetSesId(id uint64) {
	user.SesId = id
}

// ------------------------- inside -------------------------

func (user *User) destory() {
	jschedule.Stop(user.ticker)
	users.Delete(user.Uid)
}

func (user *User) tick() {
	jlog.Debug("user tick")
	if user.Base.Tick() {
		user.destory()
	}
}
