package jglobal

import (
	"jschedule"
	"sync"
	"time"
)

type tcCell[T any] struct {
	data   T
	expire int64
}

// 过期淘汰缓存
type TimeCache[T1 comparable, T2 any] struct {
	data   *LRU[T1, *tcCell[T2]]
	expire int64
	mutex  sync.RWMutex
}

// ------------------------- outside -------------------------

func NewTimeCache[T1 comparable, T2 any](expire int64) *TimeCache[T1, T2] {
	cache := &TimeCache[T1, T2]{
		data:   NewLRU[T1, *tcCell[T2]](-1),
		expire: expire,
	}
	jschedule.DoEvery(time.Minute, cache.tick)
	return cache
}

func (tc *TimeCache[T1, T2]) Set(key T1, val T2) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.data.Set(key, &tcCell[T2]{data: val, expire: time.Now().Unix() + tc.expire})
}

func (tc *TimeCache[T1, T2]) Get(key T1) T2 {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	c := tc.data.Get(key)
	if c != nil {
		c.expire = time.Now().Unix() + tc.expire
		return c.data
	}
	var zero T2
	return zero
}

func (tc *TimeCache[T1, T2]) Del(key T1) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.data.Del(key)
}

// ------------------------- inside -------------------------

func (tc *TimeCache[T1, T2]) tick(args ...any) {
	now := time.Now().Unix()
	tc.mutex.Lock()
	for {
		if c := tc.data.Peak(); c != nil && c.expire <= now {
			tc.data.Pop()
		} else {
			break
		}
	}
	tc.mutex.Unlock()
}
