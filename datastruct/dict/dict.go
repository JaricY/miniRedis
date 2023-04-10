package dict

// Consumer 是用于遍历Dict的函数，具体的由用户传入。如果返回了false则说明遍历中断
type Consumer func(key string, val interface{}) bool

// Dict 这里定义的Dict是一个接口，定义了Dict需要实现的方法
type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int
	Put(key string, val interface{}) (result int)
	PutIfAbsent(key string, val interface{}) (result int)
	PutIfExists(key string, val interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
}
