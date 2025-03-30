package juser

import (
	"jdb"
	"jglobal"
	"jmongo"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
)

type User struct {
	Uid        uint32
	live       int
	mutex      sync.Mutex
	DirtyRedis map[string]any
	DirtyMongo map[string]any
}

// ------------------------- outside -------------------------

func NewUser(uid uint32) *User {
	user := &User{
		Uid:        uid,
		DirtyRedis: map[string]any{},
		DirtyMongo: map[string]any{},
		live:       jglobal.USER_LIVE,
	}
	return user
}

func (user *User) GetUid() uint32 {
	return user.Uid
}

func (user *User) Lock() {
	user.mutex.Lock()
}

func (user *User) UnLock() {
	user.mutex.Unlock()
}

func (user *User) Touch() {
	user.live = jglobal.USER_LIVE
}

func (user *User) Flush() {
	user.flush()
}

func (user *User) Tick() bool {
	user.flush()
	if user.live -= 1; user.live <= 0 {
		return true
	}
	return false
}

// ------------------------- inside -------------------------

// 脏数据落地
func (user *User) flush() {
	user.mutex.Lock()
	if user.DirtyRedis == nil {
		user.DirtyRedis = map[string]any{}
		user.mutex.Unlock()
		jdb.Redis.Del(jglobal.Itoa(user.Uid))
	} else if len(user.DirtyRedis) > 0 {
		dirtyRedis := []any{}
		for k, v := range user.DirtyRedis {
			dirtyRedis = append(dirtyRedis, k, v)
		}
		user.DirtyRedis = map[string]any{}
		user.mutex.Unlock()
		jdb.Redis.HSet(jglobal.Itoa(user.Uid), dirtyRedis...)
	} else {
		user.mutex.Unlock()
	}
	user.mutex.Lock()
	var dirtyMongo map[string]any
	if len(user.DirtyMongo) > 0 {
		dirtyMongo, user.DirtyMongo = user.DirtyMongo, map[string]any{}
	}
	user.mutex.Unlock()
	if dirtyMongo != nil {
		a, b, update := bson.M{}, bson.M{}, bson.M{}
		for k, v := range dirtyMongo {
			if v != nil {
				a[k] = v
			} else {
				b[k] = 1
			}
		}
		if len(a) > 0 {
			update["$set"] = a
		}
		if len(b) > 0 {
			update["$unset"] = b
		}
		in := &jmongo.Input{
			Col:    jglobal.MONGO_USER,
			Filter: bson.M{"_id": user.Uid},
			Update: update,
			Upsert: true,
		}
		jdb.Mongo.UpdateOne(in)
	}
}
