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

func (que *Queue[T]) Push(data T) {
	que.buffer[que.w] = data
	que.w++
	if que.w == que.size {
		que.w = 0
	}
	if que.w == que.r {
		size0, size1 := len(que.buffer), 0
		if size0 < 1024 {
			size1 = size0 * 2
		} else {
			size1 = size0 + size0/4
		}
		buffer := make([]T, size1)
		copy(buffer, que.buffer[que.r:])
		copy(buffer[size0-que.r:], que.buffer[:que.r])
		que.buffer = buffer
		que.size = size1
		que.r = 0
		que.w = size0
	}
}

func (que *Queue[T]) Peek() (T, bool) {
	if que.r == que.w {
		var zero T
		return zero, false
	}
	return que.buffer[que.r], true
}

func (que *Queue[T]) Pop() T {
	if que.r == que.w {
		jlog.Panic("queue is empty.")
	}
	tmp := que.buffer[que.r]
	que.r++
	if que.r == que.size {
		que.r = 0
	}
	return tmp
}
