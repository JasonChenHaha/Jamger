package jglobal

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
		if _, ok := m[T2(i)]; !ok {
			*s = append(*s, v)
		} else if o != nil {
			*o = append(*o, v)
		}
	}
}
