package server

import (
	"miniRedis/database"
	"miniRedis/lib/sync/atomic"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// Handler 用于处理用户的请求,表示redis服务器的处理器
type Handler struct {

	// 表示当前活动的客户端连接，键是客户端连接的指针，值是占位符（不包含任何东西），相当于是一个Set
	activeConn sync.Map //

	// 当前活动的数据库
	db database.DB

	// 表示是否正在拒绝新的客户端请求
	closing atomic.Boolean
}
