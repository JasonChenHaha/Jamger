package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jschedule"
	"juserBase"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var users sync.Map

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	*juserBase.Base
	Uid uint32
	*Auth
	ticker any
}

// ------------------------- outside -------------------------

func Init() {}

func GetUser(uid uint32) *User {
	if user, ok := users.Load(uid); ok {
		return user.(*User)
	} else {
		in := &jmongo.Input{
			Col:     jglobal.MONGO_USER,
			Filter:  bson.M{"_id": uid},
			Project: bson.M{"auth": 1},
		}
		mData := map[string]any{}
		if err := jdb.Mongo.FindOne(in, &mData); err != nil {
			jlog.Error(err)
			return nil
		}
		rData, err := jdb.Redis.HGetAll(jglobal.Itoa(uid))
		if err != nil {
			jlog.Error(err)
			return nil
		}
		user := &User{
			Base: juserBase.NewBase(uid),
			Uid:  uid,
		}
		user.Auth = newAuth(user, mData, rData)
		user.ticker = jschedule.DoEvery(time.Second, user.tick)
		users.Store(uid, user)
		return user
	}
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
