package network

import (
	"github.com/panjf2000/gnet"
	"log"
	"miniRedis/datastruct"
	"runtime"
)

//网络处理模块，负责客户端和服务端的网络通信

var server = &RedisServer{}

// ReactHandler 处理接收到的请求
func ReactHandler(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	client := server.clients.DictFetchValue(c.Context()).(*RedisClient) //从字典中获取到client客户端对象

	client.queryBus = datastruct.NewSDS(string(frame)) // 创建查询的sds

	defer func() { //处理异常结果
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Print(err)
			log.Printf("==> %s \n", string(buf[:n]))
		}
	}()
	//client.ProcessInputBuffer() //处理请求
	out = []byte("+PONG\r\n")
	return out, action
}

// AcceptHandler 用于处理接受到新连接并创建一个客户端对象
func AcceptHandler(c gnet.Conn) (out []byte, action gnet.Action) {
	client := CreateClient(c)
	server.clients.DictAdd(client.id, client)
	return out, action
}

// ProcessInputBuffer 处理客户端接收到的数据
func (c *RedisClient) ProcessInputBuffer() {

}

func CreateClient(c gnet.Conn) *RedisClient {
	c.SetContext(GenerateClientId())
	return &RedisClient{
		id:     c.Context().(int),
		conn:   c,
		db:     server.db,
		argc:   0,
		buf:    make([]byte, 1024*12),
		bufpos: 0,
	}
}

func GenerateClientId() int {
	server.clientCounter++
	return server.clientCounter
}
