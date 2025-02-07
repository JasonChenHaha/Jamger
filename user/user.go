package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jschedule"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	EXPIRE = 60 * 10
)

// 所有属性的写需要使用对应的set方法，以驱动数据定时落地
type User struct {
	Uid        uint32
	AesKey     []byte
	expire     int
	dirtyRedis map[string]any
	dirtyMongo map[string]any
	mutex      sync.Mutex
}

// ------------------------- package -------------------------

func newUser(uid uint32) *User {
	user := &User{
		Uid:    uid,
		expire: EXPIRE,
	}
	res, err := jdb.Redis.HGetAll(jglobal.Itoa(uid))
	if err != nil {
		jlog.Error(err)
		return nil
	}
	if len(res) > 0 {
		user.AesKey = []byte(res["aesKey"])
	}
	jschedule.DoEvery(time.Second, user.tick)
	return user
}

func (user *User) refresh() {
	user.expire = EXPIRE
}

// ------------------------- outside -------------------------

func (user *User) Lock() {
	user.mutex.Lock()
}

func (user *User) Unlock() {
	user.mutex.Unlock()
}

func (user *User) SetAesKey(key []byte) {
	user.AesKey = key
	user.dirtyRedis["aesKey"] = key
}

// ------------------------- inside -------------------------

func (user *User) tick() {
	dirtyRedis := []any{}
	user.mutex.Lock()
	for k, v := range user.dirtyRedis {
		dirtyRedis = append(dirtyRedis, k, v)
	}
	dirtyMongo := bson.M(user.dirtyMongo)
	// to do: 减少内存碎片
	user.dirtyRedis = map[string]any{}
	user.dirtyMongo = map[string]any{}
	user.mutex.Unlock()
	if _, err := jdb.Redis.HSet(jglobal.Itoa(user.Uid), dirtyRedis...); err != nil {
		jlog.Error(err)
	}
	in := &jmongo.Input{
		Col:    "account",
		Filter: bson.M{"_id": user.Uid},
		Update: bson.M{"$set": dirtyMongo},
	}
	if err := jdb.Mongo.UpdateOne(in); err != nil {
		jlog.Error(err)
	}
	if user.expire -= 1; user.expire <= 0 {
		delete(user.Uid)
	}
}
