package jglobal

import (
	"sync"
	"time"
)

// 过期淘汰缓存
type TimeCache[T1 comparable, T2 any] struct {
	data   map[T1]T2
	ts     map[T1]int64
	expire int64
	mutex  sync.RWMutex
}

// ------------------------- outside -------------------------

func NewTimeCache[T1 comparable, T2 any](expire int64) *TimeCache[T1, T2] {
	return &TimeCache[T1, T2]{
		data:   map[T1]T2{},
		ts:     map[T1]int64{},
		expire: expire,
	}
}

func (tc *TimeCache[T1, T2]) Set(key T1, val T2) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.data[key] = val
	tc.ts[key] = time.Now().Unix() + tc.expire
}

func (tc *TimeCache[T1, T2]) Get(key T1) T2 {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	if t, ok := tc.ts[key]; !ok || t <= time.Now().Unix() {
		var zero T2
		return zero
	}
	return tc.data[key]
}
