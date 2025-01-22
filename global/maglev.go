package jglobal

import (
	"strconv"

	"github.com/dchest/siphash"
)

const THE_NUM = 65537

type Maglev[T comparable] struct {
	lookup []T
}

// ------------------------- outside -------------------------

func NewMaglev[T comparable](node map[int]T) *Maglev[T] {
	m := &Maglev[T]{lookup: make([]T, THE_NUM)}
	m.genLookupTable(node)
	return m
}

func (ml *Maglev[T]) Get(key any) T {
	var id uint64
	switch o := key.(type) {
	case string:
		id = siphash.Hash(0, 0, []byte(o))
	case int:
		id = uint64(o)
	case uint:
		id = uint64(o)
	case int16:
		id = uint64(o)
	case uint16:
		id = uint64(o)
	case int32:
		id = uint64(o)
	case uint32:
		id = uint64(o)
	case int64:
		id = uint64(o)
	case uint64:
		id = uint64(o)
	}
	return ml.lookup[id%THE_NUM]
}

// ------------------------- inside -------------------------

func (ml *Maglev[T]) genLookupTable(node map[int]T) {
	permutation := map[int][]uint64{}
	for k := range node {
		permutation[k] = make([]uint64, THE_NUM)
		for i := 0; i < THE_NUM; i++ {
			by := []byte(strconv.Itoa(k))
			offset := siphash.Hash(0, 0, by) % THE_NUM
			ship := siphash.Hash(1, 1, by)%(THE_NUM-1) + 1
			permutation[k][i] = (offset + uint64(i)*ship) % THE_NUM
		}
	}
	var zero T
	n := uint64(0)
	for {
		for k := range permutation {
			for len(permutation[k]) > 0 {
				idx := permutation[k][0]
				permutation[k] = permutation[k][1:]
				if ml.lookup[idx] == zero {
					ml.lookup[idx] = node[k]
					n++
					break
				}
			}
			if len(permutation[k]) == 0 {
				delete(permutation, k)
			}
		}
		if len(permutation) == 0 || n == THE_NUM {
			break
		}
	}
}
