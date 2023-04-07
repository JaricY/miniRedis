package network

import (
	"github.com/panjf2000/gnet"
	"miniRedis/datastruct"
	"miniRedis/db"
)

const (
	redisReqInline    = 1
	redisReqMultibulk = 2
)

type RedisClient struct {
	Id   int
	Conn gnet.Conn   //客户端连接
	Db   *db.RedisDb //当前客户端使用的数据库编号
	name *datastruct.RedisObj
	Argc int                    //命令数量
	Argv []*datastruct.RedisObj //命令值

	Cmd     *RedisCommand //当前执行的命令
	LastCmd *RedisCommand //最后执行的命令

	ReqType  int
	QueryBus *datastruct.SDS // 从客户端的请求获取到的数据

	Buf     []byte //准备发回客户端的数据
	Bufpos  int    //发回给客户端的数据pos末尾点
	SentLen int    //已发送的字节数，也就是需要发送下一个数据的起始点
}
type RedisCommand struct {
	name             datastruct.SDS            //命令名字
	redisCommandFunc func(client *RedisClient) //命令解析方法
}

func CreateClient(c gnet.Conn) *RedisClient {
	c.SetContext(GenerateClientId())
	return &RedisClient{
		Id:     c.Context().(int),
		Conn:   c,
		Db:     server.db,
		Argc:   0,
		Buf:    make([]byte, 1024*12),
		Bufpos: 0,
	}
}
