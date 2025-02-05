package jglobal

// 自扩容环形缓存

type Circle[T any] struct {
	head, tail int
	size       int
	cap        int
	data       []T
}

// ------------------------- outside -------------------------

func NewCircle[T any]() *Circle[T] {
	return &Circle[T]{
		cap:  8,
		data: make([]T, 8),
	}
}

func (o *Circle[T]) Size() int {
	return o.size
}

func (o *Circle[T]) Push(cell T) {
	if o.head == o.tail && o.size == o.cap {
		cap := 0
		if o.cap < 1024 {
			cap = o.cap * 2
		} else {
			cap = o.cap + o.cap/4
		}
		tmp := make([]T, cap)
		copy(tmp, o.data[o.tail:o.cap])
		copy(tmp[o.cap-o.tail:], o.data[0:o.head])
		o.data = tmp
		o.tail = 0
		o.head = o.cap
		o.cap = cap
	}
	o.data[o.head] = cell
	o.head = (o.head + 1) % o.cap
	o.size++
}

func (o *Circle[T]) Pop() T {
	if o.tail == o.head && o.size == 0 {
		var zero T
		return zero
	}
	tmp := o.data[o.tail]
	o.tail = (o.tail + 1) % (o.size + 1)
	o.size--
	return tmp
}

func (o *Circle[T]) Tail() T {
	if o.size == 0 {
		var zero T
		return zero
	}
	return o.data[o.tail]
}
