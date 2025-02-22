package jglobal

// ------------------------- outside -------------------------

// 查找元素
func SliceFind[T comparable](s []T, val T) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}

// 获取多个下标内的元素
func SliceMultiPick[T1 any, T2 AllInt](s []T1, o []T1, idx []T2) {
	for _, v := range idx {
		o = append(o, s[v])
	}
}

// 删除下标
func SliceDeletePos[T1 any, T2 AllInt](s *[]T1, idx T2) {
	*s = append((*s)[:idx], (*s)[idx+1:]...)
}

// 删除多个下标
func SliceDeleteMultiPos[T1 any, T2 AllInt](s *[]T1, idx []T2, o *[]T1) {
	m := map[T2]any{}
	for _, v := range idx {
		m[v] = struct{}{}
	}
	tmp := *s
	(*s) = (*s)[:0]
	for i, v := range tmp {
		if m[T2(i)] == nil {
			*s = append(*s, v)
		} else if o != nil {
			*o = append(*o, v)
		}
	}
}

func NewHashSlice[T1 AllIntString, T2 comparable]() *HashSlice[T1, T2] {
	return &HashSlice[T1, T2]{
		hash:  map[T1]T2{},
		slice: []T2{},
	}
}

// 哈希和切片结合
// 适用既需要快速访问, 又需要有序访问的场景
type HashSlice[T1 AllIntString, T2 comparable] struct {
	hash  map[T1]T2
	slice []T2
}

func (hs *HashSlice[T1, T2]) Insert(key T1, data T2) {
	hs.hash[key] = data
	hs.slice = append(hs.slice, data)
}

func (hs *HashSlice[T1, T2]) Len() int {
	return len(hs.slice)
}

func (hs *HashSlice[T1, T2]) Get(key T1) T2 {
	return hs.hash[key]
}

func (hs *HashSlice[T1, T2]) IndexOf(idx int) T2 {
	if idx < len(hs.slice) {
		return hs.slice[idx]
	}
	var zero T2
	return zero
}

func (hs *HashSlice[T1, T2]) Keys() []T1 {
	i, o := 0, make([]T1, len(hs.hash))
	for k := range hs.hash {
		o[i] = k
		i++
	}
	return o
}

func (hs *HashSlice[T1, T2]) Values() []T2 {
	return hs.slice
}

func (hs *HashSlice[T1, T2]) KeyValues() map[T1]T2 {
	return hs.hash
}

func (hs *HashSlice[T1, T2]) Del(key T1) {
	if _, ok := hs.hash[key]; !ok {
		return
	}
	for k, v := range hs.slice {
		if hs.hash[key] == v {
			hs.slice = append(hs.slice[:k], hs.slice[k+1:]...)
			break
		}
	}
	delete(hs.hash, key)
}
