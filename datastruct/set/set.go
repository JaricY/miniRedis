package set

import "miniRedis/datastruct/dict"

// Set 是一组基于哈希表的元素
type Set struct {
	dict dict.Dict
}

func Make(members ...string) *Set {
	set := &Set{
		dict: dict.MakeSimple(),
	}
	for _, member := range members {
		set.Add(member)
	}
	return set
}

func (set *Set) Add(val string) int {
	return set.dict.Put(val, nil)
}

func (set *Set) Remove(val string) int {
	return set.dict.Remove(val)
}

func (set *Set) Has(val string) bool {
	_, exists := set.dict.Get(val)
	return exists
}

func (set *Set) Len() int {
	return set.dict.Len()
}

// ToSlice 将集合中的元素转换成切片
func (set *Set) ToSlice() []string {
	slice := make([]string, set.Len())
	i := 0
	set.dict.ForEach(func(key string, val interface{}) bool {
		if i < len(slice) {
			slice[i] = key
		} else {
			// set extended during traversal
			slice = append(slice, key)
		}
		i++
		return true
	})
	return slice
}

func (set *Set) ForEach(consumer func(member string) bool) {
	set.dict.ForEach(func(key string, val interface{}) bool {
		return consumer(key)
	})
}

// Intersect 返回两个集合的交集
func (set *Set) Intersect(another *Set) *Set {
	if set == nil {
		panic("set is nil")
	}

	result := Make()
	another.ForEach(func(member string) bool {
		if set.Has(member) {
			result.Add(member)
		}
		return true
	})
	return result
}

// Union 返回两个集合的并集
func (set *Set) Union(another *Set) *Set {
	if set == nil {
		panic("set is nil")
	}
	result := Make()
	another.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})
	set.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})
	return result
}

// Diff 求set集合中减去another集合中相同的元素
func (set *Set) Diff(another *Set) *Set {
	if set == nil {
		panic("set is nil")
	}

	result := Make()
	set.ForEach(func(member string) bool {
		if !another.Has(member) {
			result.Add(member)
		}
		return true
	})
	return result
}

func (set *Set) RandomMembers(limit int) []string {
	return set.dict.RandomKeys(limit)
}

func (set *Set) RandomDistinctMembers(limit int) []string {
	return set.dict.RandomDistinctKeys(limit)
}
