package jglobal

import (
	"jlog"
)

type Queue[T any] struct {
	buffer []T
	size   int
	r      int
	w      int
}

// ------------------------- outside -------------------------

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		buffer: make([]T, 8),
		size:   8,
	}
}

func (o *Queue[T]) Push(data T) {
	o.buffer[o.w] = data
	o.w++
	if o.w == o.size {
		o.w = 0
	}
	if o.w == o.r {
		size0, size1 := len(o.buffer), 0
		if size0 < 1024 {
			size1 = size0 * 2
		} else {
			size1 = size0 + size0/4
		}
		buffer := make([]T, size1)
		copy(buffer, o.buffer[o.r:])
		copy(buffer[size0-o.r:], o.buffer[:o.r])
		o.buffer = buffer
		o.size = size1
		o.r = 0
		o.w = size0
	}
}

func (o *Queue[T]) Peek() (T, bool) {
	if o.r == o.w {
		var zero T
		return zero, false
	}
	return o.buffer[o.r], true
}

func (o *Queue[T]) Pop() T {
	if o.r == o.w {
		jlog.Panic("queue is empty.")
	}
	tmp := o.buffer[o.r]
	o.r++
	if o.r == o.size {
		o.r = 0
	}
	return tmp
}
