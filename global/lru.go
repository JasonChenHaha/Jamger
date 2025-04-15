package jglobal

type lruCell[T1 comparable, T2 any] struct {
	prev *lruCell[T1, T2]
	next *lruCell[T1, T2]
	key  T1
	data T2
}

type LRU[T1 comparable, T2 any] struct {
	data map[T1]*lruCell[T1, T2]
	head *lruCell[T1, T2]
	tail *lruCell[T1, T2]
	size int
}

// ------------------------- outside -------------------------

func NewLRU[T1 comparable, T2 any](size int) *LRU[T1, T2] {
	lru := &LRU[T1, T2]{
		data: map[T1]*lruCell[T1, T2]{},
		head: &lruCell[T1, T2]{},
		tail: &lruCell[T1, T2]{},
		size: size,
	}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

func (lru *LRU[T1, T2]) Set(key T1, val T2) {
	c := lru.data[key]
	if c != nil {
		c.prev.next = c.next
		c.next.prev = c.prev
		c.data = val
	} else {
		c = &lruCell[T1, T2]{key: key, data: val}
		lru.data[key] = c
		if lru.size > 0 && lru.size < len(lru.data) {
			x := lru.head.next
			x.prev.next = x.next
			x.next.prev = x.prev
			delete(lru.data, x.key)
		}
	}
	lru.tail.prev.next = c
	c.prev = lru.tail.prev
	c.next = lru.tail
	lru.tail.prev = c
}

func (lru *LRU[T1, T2]) Get(key T1) T2 {
	if c := lru.data[key]; c != nil {
		c.prev.next = c.next
		c.next.prev = c.prev
		lru.tail.prev.next = c
		c.prev = lru.tail.prev
		c.next = lru.tail
		lru.tail.prev = c
		return c.data
	} else {
		var zero T2
		return zero
	}
}

func (lru *LRU[T1, T2]) Del(key T1) {
	if c := lru.data[key]; c != nil {
		c.prev.next = c.next
		c.next.prev = c.prev
		delete(lru.data, key)
	}
}

// 窥探头部节点
func (lru *LRU[T1, T2]) Peak() T2 {
	if lru.head.next != lru.tail {
		return lru.head.next.data
	}
	var zero T2
	return zero
}

// 删除头部节点
func (lru *LRU[T1, T2]) Pop() {
	if lru.head.next == lru.head {
		return
	}
	c := lru.head.next
	c.prev.next = c.next
	c.next.prev = c.prev
	delete(lru.data, c.key)
}
