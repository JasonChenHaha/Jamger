package jglobal

import "sync"

type Queue[T any] struct {
	Data  []T
	mutex sync.RWMutex
}

// ------------------------- outside -------------------------

func (que *Queue[T]) Push(data T) {
	que.mutex.Lock()
	defer que.mutex.Unlock()
	que.Data = append(que.Data, data)
}

func (que *Queue[T]) Pick() T {
	que.mutex.Lock()
	defer que.mutex.Unlock()
	if len(que.Data) == 0 {
		var zero T
		return zero
	}
	data := que.Data[0]
	que.Data = que.Data[1:]
	return data
}

func (que *Queue[T]) PickAll() any {
	que.mutex.Lock()
	defer que.mutex.Unlock()
	data := que.Data
	que.Data = nil
	return data
}
