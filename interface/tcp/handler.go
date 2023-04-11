package tcp

import (
	"context"
	"net"
)

// HandleFunc 代表处理方法，ctx表示请求携带的相关数据，conn表示一个网络连接，用于客户端和服务端之间传递数据
type HandleFunc func(ctx context.Context, conn net.Conn)

// Handler 用于处理tcp的服务
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
