package juserBase

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
)

type User interface {
	SetAesKey([]byte)
}

const (
	EXPIRE = 10 // 60 * 10
)

type Base struct {
	key         string
	DirtyRedis  map[string]any
	dirtyRedis2 []any
	DirtyMongo  map[string]any
	expire      int
	mutex       sync.Mutex
}

// ------------------------- outside -------------------------

func NewBase(key uint32) *Base {
	base := &Base{
		key:         jglobal.Itoa(key),
		DirtyRedis:  map[string]any{},
		dirtyRedis2: []any{},
		DirtyMongo:  map[string]any{},
		expire:      EXPIRE,
	}
	return base
}

func (base *Base) Lock() {
	base.mutex.Lock()
}

func (base *Base) Unlock() {
	base.mutex.Unlock()
}

func (base *Base) Refresh() {
	base.expire = EXPIRE
}

func (base *Base) Flush() {
	base.flush(false)
}

func (base *Base) Tick() bool {
	base.flush(true)
	if base.expire -= 1; base.expire <= 0 {
		return true
	}
	return false
}

// ------------------------- inside -------------------------

// 脏数据落地
func (base *Base) flush(needLock bool) error {
	sizeRedis, sizeMongo := len(base.DirtyRedis), len(base.DirtyMongo)
	if sizeRedis > 0 || sizeMongo > 0 {
		if needLock {
			if ok := base.mutex.TryLock(); !ok {
				return nil
			}
		}
		if sizeRedis > 0 {
			base.dirtyRedis2 = base.dirtyRedis2[:0]
			for k, v := range base.DirtyRedis {
				base.dirtyRedis2 = append(base.dirtyRedis2, k, v)
			}
			base.DirtyRedis = map[string]any{}
		}
		var dirtyMongo map[string]any
		if sizeMongo > 0 {
			dirtyMongo, base.DirtyMongo = base.DirtyMongo, map[string]any{}
		}
		if needLock {
			base.mutex.Unlock()
		}
		if len(base.dirtyRedis2) > 0 {
			if _, err := jdb.Redis.HSet(base.key, base.dirtyRedis2...); err != nil {
				jlog.Error(err)
				return err
			}
		}
		if len(dirtyMongo) > 0 {
			in := &jmongo.Input{
				Col:    jglobal.MONGO_USER,
				Filter: bson.M{"_id": base.key},
				Update: bson.M{"$set": dirtyMongo},
				Upsert: true,
			}
			if err := jdb.Mongo.UpdateOne(in); err != nil {
				jlog.Error(err)
				return err
			}
		}
	}
	return nil
}
