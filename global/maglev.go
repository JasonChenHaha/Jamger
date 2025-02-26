package jglobal

import (
	"slices"
	"sort"
	"strconv"

	"github.com/dchest/siphash"
)

type Maglev[T comparable] struct {
	lookup []T
}

const THE_NUM = 65537

// ------------------------- outside -------------------------

func NewMaglev[T comparable](node map[int]T) *Maglev[T] {
	m := &Maglev[T]{lookup: make([]T, THE_NUM)}
	m.genLookupTable(node)
	return m
}

func (o *Maglev[T]) Get(key any) T {
	var id uint64
	switch v := key.(type) {
	case string:
		id = siphash.Hash(0, 0, []byte(v))
	case int:
		id = uint64(v)
	case uint:
		id = uint64(v)
	case int16:
		id = uint64(v)
	case uint16:
		id = uint64(v)
	case int32:
		id = uint64(v)
	case uint32:
		id = uint64(v)
	case int64:
		id = uint64(v)
	case uint64:
		id = uint64(v)
	}
	return o.lookup[id%THE_NUM]
}

// ------------------------- inside -------------------------

func (o *Maglev[T]) genLookupTable(node map[int]T) {
	permutation := map[int][]uint64{}
	keys := []int{}
	for k := range node {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, v := range keys {
		permutation[v] = make([]uint64, THE_NUM)
		for i := range THE_NUM {
			by := []byte(strconv.Itoa(v))
			offset := siphash.Hash(0, 0, by) % THE_NUM
			ship := siphash.Hash(1, 1, by)%(THE_NUM-1) + 1
			permutation[v][i] = (offset + uint64(i)*ship) % THE_NUM
		}
	}
	var zero T
	n := uint64(0)
	for {
		for i := len(keys) - 1; i >= 0; i-- {
			v := keys[i]
			for len(permutation[v]) > 0 {
				idx := permutation[v][0]
				permutation[v] = permutation[v][1:]
				if o.lookup[idx] == zero {
					o.lookup[idx] = node[v]
					n++
					break
				}
			}
			if len(permutation[v]) == 0 {
				keys = slices.Delete(keys, i, i+1)
			}
		}
		if len(permutation) == 0 || n == THE_NUM {
			break
		}
	}
}
