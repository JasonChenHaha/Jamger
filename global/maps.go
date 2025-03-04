package jglobal

import "sync"

type Maps[T AllInt] struct {
	matrix []sync.Map
	size   T
}

// ------------------------- outside -------------------------

func NewMaps[T AllInt](size T) *Maps[T] {
	return &Maps[T]{
		matrix: make([]sync.Map, size),
		size:   size,
	}
}

func (m *Maps[T]) Load(key T) (any, bool) {
	return m.matrix[key%m.size].Load(key)
}

func (m *Maps[T]) Store(key T, val any) {
	m.matrix[key%m.size].Store(key, val)
}

func (m *Maps[T]) Delete(key T) {
	m.matrix[key%m.size].Delete(key)
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
