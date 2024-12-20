package jglobal

import (
	"jlog"

	"github.com/dchest/siphash"
)

const THE_NUM = 65537

type Maglev struct {
	node   []string
	lookup []string
}

// ------------------------- outside -------------------------

func NewMaglev(node []string) *Maglev {
	m := &Maglev{
		node:   node,
		lookup: make([]string, THE_NUM),
	}
	m.genLookupTable()
	return m
}

func (ml *Maglev) Get(key any) string {
	var id uint64
	switch o := key.(type) {
	case string:
		id = siphash.Hash(0, 0, []byte(o))
	case int:
	case uint:
	case int16:
	case uint16:
	case int32:
	case uint32:
	case int64:
	case uint64:
		id = uint64(o)
	}
	return ml.lookup[id%THE_NUM]
}

// ------------------------- inside -------------------------

func (ml *Maglev) genLookupTable() {
	permutation := map[string][]uint64{}
	for _, v := range ml.node {
		permutation[v] = make([]uint64, THE_NUM)
		for i := 0; i < THE_NUM; i++ {
			by := []byte(v)
			offset := siphash.Hash(0, 0, by) % THE_NUM
			ship := siphash.Hash(1, 1, by)%(THE_NUM-1) + 1
			permutation[v][i] = (offset + uint64(i)*ship) % THE_NUM
		}
	}
	n := uint64(0)
	for {
		for k := range permutation {
			for len(permutation[k]) > 0 {
				idx := permutation[k][0]
				permutation[k] = permutation[k][1:]
				if ml.lookup[idx] == "" {
					ml.lookup[idx] = k
					n++
					break
				}
			}
			if len(permutation[k]) == 0 {
				delete(permutation, k)
			}
		}
		if n == THE_NUM {
			break
		}
	}
	// for test
	s := map[string]int{}
	for _, v := range ml.lookup {
		s[v]++
	}
	jlog.Debug(s)
}
