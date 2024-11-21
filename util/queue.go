package util

import "sync"

type Queue struct {
	Data []any
	lock sync.RWMutex
}

func (que *Queue) Push(data any) {
	que.lock.Lock()
	defer que.lock.Unlock()

	que.Data = append(que.Data, data)
}

func (que *Queue) Pick() any {
	que.lock.Lock()
	defer que.lock.Unlock()

	if len(que.Data) == 0 {
		return nil
	}
	data := que.Data[0]
	que.Data = que.Data[1:]
	return data
}

func (que *Queue) PickAll() any {
	que.lock.Lock()
	defer que.lock.Unlock()

	data := que.Data
	que.Data = nil
	return data
}
