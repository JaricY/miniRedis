package pubsub

import (
	"miniRedis/datastruct/dict"
	"miniRedis/datastruct/lock"
)

// Hub stores all subscribe relations
type Hub struct {
	// 存储频道与订阅者的关系，键是频道，值是订阅者
	subs dict.Dict
	// 用于多协程同步读写频道内容
	subsLocker *lock.Locks
}

// MakeHub creates new hub
func MakeHub() *Hub {
	return &Hub{
		subs:       dict.MakeConcurrent(4),
		subsLocker: lock.Make(16),
	}
}
