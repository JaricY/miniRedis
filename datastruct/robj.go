package datastruct

type RedisType uint8     //对象类型
type RedisEncoding uint8 //编码方式

const (
	RedisTypeString    RedisType = iota
	RedisTypeHash      RedisType = 1
	RedisTypeList      RedisType = 2
	RedisTypeSet       RedisType = 3
	RedisTypeSortedSet RedisType = 4
)

const (
	RedisEncodingInt    RedisEncoding = iota
	RedisEncodingString RedisEncoding = 1
	// ...
)

type RedisObj struct {
	Type     RedisType
	Encoding RedisEncoding
	LRU      int64
	RefCount int //引用计数
	Ptr      interface{}
}

func CreateObject(t RedisType, ptr interface{}) *RedisObj {
	return &RedisObj{
		Type:     t,
		Encoding: 0,
		RefCount: 1,
		Ptr:      ptr,
		//LRU:      lruClock(),
	}
}
