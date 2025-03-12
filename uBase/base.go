package juBase

import (
	"jdb"
	"jglobal"
	"jmongo"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
)

type Base struct {
	Uid         uint32
	live        int
	mutex       sync.Mutex
	DirtyRedis  map[string]any
	DirtyMongo  map[string]any
	DirtyMongo2 map[string]any
}

// ------------------------- outside -------------------------

func NewBase(uid uint32) *Base {
	base := &Base{
		Uid:         uid,
		DirtyRedis:  map[string]any{},
		DirtyMongo:  map[string]any{},
		DirtyMongo2: map[string]any{},
		live:        jglobal.USER_LIVE,
	}
	return base
}

func (base *Base) GetUid() uint32 {
	return base.Uid
}

func (base *Base) Lock() {
	base.mutex.Lock()
}

func (base *Base) UnLock() {
	base.mutex.Unlock()
}

func (base *Base) Touch() {
	base.live = jglobal.USER_LIVE
}

func (base *Base) Flush() {
	base.flush()
}

func (base *Base) Tick() bool {
	base.flush()
	if base.live -= 1; base.live <= 0 {
		return true
	}
	return false
}

// ------------------------- inside -------------------------

// 脏数据落地
func (base *Base) flush() {
	base.mutex.Lock()
	if base.DirtyRedis == nil {
		base.DirtyRedis = map[string]any{}
		base.mutex.Unlock()
		jdb.Redis.Del(jglobal.Itoa(base.Uid))
	} else if len(base.DirtyRedis) > 0 {
		dirtyRedis := []any{}
		for k, v := range base.DirtyRedis {
			dirtyRedis = append(dirtyRedis, k, v)
		}
		base.DirtyRedis = map[string]any{}
		base.mutex.Unlock()
		jdb.Redis.HSet(jglobal.Itoa(base.Uid), dirtyRedis...)
	} else {
		base.mutex.Unlock()
	}
	base.mutex.Lock()
	var dirtyMongo, dirtyMongo2 map[string]any
	if len(base.DirtyMongo) > 0 || len(base.DirtyMongo2) > 0 {
		dirtyMongo, base.DirtyMongo = base.DirtyMongo, map[string]any{}
		dirtyMongo2, base.DirtyMongo2 = base.DirtyMongo2, map[string]any{}
	}
	base.mutex.Unlock()
	if dirtyMongo != nil {
		in := &jmongo.Input{
			Col:    jglobal.MONGO_USER,
			Filter: bson.M{"_id": base.Uid},
			Update: bson.M{"$set": dirtyMongo, "$unset": dirtyMongo2},
			Upsert: true,
		}
		jdb.Mongo.UpdateOne(in)
	}
}
