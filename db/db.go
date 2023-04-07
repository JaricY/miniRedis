package db

import (
	"minRedis/datastruct"
)

type RedisDb struct { //对应的数据库，从0号开始
	id     int              // 序号
	dict   *datastruct.Dict // 用于存储数据
	expire *datastruct.Dict // 用于存储过期时间
}

func (r *RedisDb) LookupKey(key interface{}) bool {
	value := r.dict.DictFetchValue(key)
	if value == nil {
		return false
	}
	return true
}

func (r *RedisDb) DbAdd(key, val interface{}) bool {
	add := r.dict.DictAdd(key, val)
	return add
}

func (r *RedisDb) DbDelete(key interface{}) {
	r.dict.DictDelete(key)
}

func (r *RedisDb) SetKey(key, val interface{}) {
	if r.LookupKey(key) {
		r.DbOverwrite(key, val)
	} else {
		r.DbAdd(key, val)
	}
}

func (r *RedisDb) DbOverwrite(key, val interface{}) {
	r.dict.DictReplace(key, val)
}

func (r *RedisDb) SetExpire(key interface{}, expire uint64) {
	if r.LookupKey(key) {
		r.expire.DictAdd(key, expire)
	}
}

func DBCreate(id int) *RedisDb {
	db := &RedisDb{
		id:     id,
		dict:   datastruct.DictCreate(datastruct.MyDictType{}, nil),
		expire: datastruct.DictCreate(datastruct.MyDictType{}, nil),
	}

	return db
}
