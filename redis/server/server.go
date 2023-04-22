package server

import (
	"context"
	"io"
	"miniRedis/config"
	database2 "miniRedis/database"
	"miniRedis/interface/database"
	"miniRedis/lib/logger"
	"miniRedis/lib/sync/atomic"
	"miniRedis/redis/connection"
	"miniRedis/redis/parser"
	"miniRedis/redis/protocol"
	"net"
	"strings"
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

func MakeHandler() *Handler {
	var db database.DB
	if config.Properties.Self != "" &&
		len(config.Properties.Peers) > 0 {
		//db = cluster.MakeCluster()
	} else {
		db = database2.NewStandaloneServer()
	}
	return &Handler{
		db: db,
	}
}

func (h *Handler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle receives and executes redis commands
func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// closing handler refuse new connection
		_ = conn.Close()
		return
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, struct{}{})

	// 解析完成后得到的管道
	ch := parser.ParseStream(conn)

	for payload := range ch {
		// 处理错误结果
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := protocol.MakeErrReply(payload.Err.Error())
			_, err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}

		//处理空结果
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}

		// 如果是多行结果，也就是*开头的
		r, ok := payload.Data.(*protocol.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk protocol")
			continue
		}

		// 执行参数
		result := h.db.Exec(client, r.Args)
		if result != nil {
			// 写回响应
			_, _ = client.Write(result.ToBytes())
		} else {
			_, _ = client.Write(unknownErrReplyBytes)
		}
	}
}
