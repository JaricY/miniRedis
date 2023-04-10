package dict

// SimpleDict 包装了一个map映射，不是线程安全的
type SimpleDict struct {
	m map[string]interface{}
}

func MakeSimple() *SimpleDict {
	return &SimpleDict{
		m: make(map[string]interface{}),
	}
}

// Get 从字典中获取元素
func (dict *SimpleDict) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m[key]
	return val, ok
}

// Len 获取当前字典的所有元素个数
func (dict *SimpleDict) Len() int {
	if dict.m == nil {
		panic("m is nil")
	}
	return len(dict.m)
}

// Put 插入一个k-v到字典中，如果已经存在则返回0，是新的则返回1
func (dict *SimpleDict) Put(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	dict.m[key] = val
	if existed {
		return 0
	}
	return 1
}

// PutIfAbsent 插入一个k-v到字典中，如果已经存在则返回0且取消插入，是新的则返回1
func (dict *SimpleDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	if existed {
		return 0
	}
	dict.m[key] = val
	return 1
}

// PutIfExists 插入一个k-v到字典中，如果已经存在则返回1，是新的则返回0，相当于是修改
func (dict *SimpleDict) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	if existed {
		dict.m[key] = val
		return 1
	}
	return 0
}

// Remove 移除key
func (dict *SimpleDict) Remove(key string) (result int) {
	_, existed := dict.m[key]
	delete(dict.m, key)
	if existed {
		return 1
	}
	return 0
}

// Keys 返回所有的key
func (dict *SimpleDict) Keys() []string {
	result := make([]string, len(dict.m))
	i := 0
	for k := range dict.m {
		result[i] = k
		i++
	}
	return result
}

// ForEach 遍历
func (dict *SimpleDict) ForEach(consumer Consumer) {
	for k, v := range dict.m {
		if !consumer(k, v) {
			break
		}
	}
}

// RandomKeys randomly 返回指定数量的随机键，可能有重复的健
func (dict *SimpleDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		for k := range dict.m {
			result[i] = k
			break
		}
	}
	return result
}

// RandomDistinctKeys 返回指定数量的不重复的随机键
func (dict *SimpleDict) RandomDistinctKeys(limit int) []string {
	size := limit
	if size > len(dict.m) {
		size = len(dict.m)
	}
	result := make([]string, size)
	i := 0
	for k := range dict.m {
		if i == size {
			break
		}
		result[i] = k
		i++
	}
	return result
}

// Clear 清空字典
func (dict *SimpleDict) Clear() {
	*dict = *MakeSimple()
}
