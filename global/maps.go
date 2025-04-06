package jglobal

// 有锁多sync.map容器，解决单sync.map排队加锁瓶颈

import (
	"sync"
	"sync/atomic"
)

type Maps[T AllInt] struct {
	matrix []sync.Map
	// for test
	count []atomic.Int32
	size  T
}

// ------------------------- outside -------------------------

func (m *Maps[T]) String() string {
	c := int32(0)
	for _, v := range m.count {
		c += v.Load()
	}
	return Itoa(c)
}

func NewMaps[T AllInt](size T) *Maps[T] {
	return &Maps[T]{
		matrix: make([]sync.Map, size),
		count:  make([]atomic.Int32, size),
		size:   size,
	}
}

func (m *Maps[T]) Load(key T) (any, bool) {
	return m.matrix[key%m.size].Load(key)
}

func (m *Maps[T]) Store(key T, val any) {
	m.matrix[key%m.size].Store(key, val)
	m.count[key%m.size].Add(1)
}

func (m *Maps[T]) Delete(key T) {
	m.matrix[key%m.size].Delete(key)
	m.count[key%m.size].Add(-1)
}

func (m *Maps[T]) Range(fun func(any, any) bool) {
	keep := true
	fun2 := func(k, v any) bool {
		keep = fun(k, v)
		return keep
	}
	for k := range m.matrix {
		m.matrix[k].Range(fun2)
		if !keep {
			return
		}
	}
}
