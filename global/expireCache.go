package jglobal

import (
	"jschedule"
	"time"
)

type ExpireCache[T any] struct {
	expire   int64
	curTm    int64
	timeline *Circle[int64]
	data     map[int64]map[uint32]T
}

// ------------------------- outside -------------------------

func NewExpireCache[T any](expire int64, interval int) *ExpireCache[T] {
	ec := &ExpireCache[T]{
		expire:   expire,
		timeline: NewCircle[int64](),
		data:     map[int64]map[uint32]T{},
	}
	jschedule.DoEvery(time.Duration(interval)*time.Second, ec.tick)
	return ec
}

func (ec *ExpireCache[T]) Add(key uint32, cell T) {
	tm := time.Now().Unix() + int64(ec.expire)
	ec.data[tm][key] = cell

}

// ------------------------- inside -------------------------

func (ec *ExpireCache[T]) tick() {
	ec.curTm = time.Now().Unix()
	if ec.timeline.Size() == 0 {
		return
	}
	tm := ec.timeline.Pop()
	delete(ec.data, tm)
}
