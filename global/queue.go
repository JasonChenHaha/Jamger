package jglobal

import "sync"

type Queue[T any] struct {
	Data []T
	lock sync.RWMutex
}

func (que *Queue[T]) Push(data T) {
	que.lock.Lock()
	defer que.lock.Unlock()
	que.Data = append(que.Data, data)
}

func (que *Queue[T]) Pick() T {
	que.lock.Lock()
	defer que.lock.Unlock()
	if len(que.Data) == 0 {
		var zero T
		return zero
	}
	data := que.Data[0]
	que.Data = que.Data[1:]
	return data
}

func (que *Queue[T]) PickAll() any {
	que.lock.Lock()
	defer que.lock.Unlock()
	data := que.Data
	que.Data = nil
	return data
}
