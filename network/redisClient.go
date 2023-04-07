package network

import (
	"github.com/panjf2000/gnet"
	"miniRedis/datastruct"
	"miniRedis/db"
)

type RedisClient struct {
	id   int
	conn gnet.Conn   //客户端连接
	db   *db.RedisDb //当前客户端使用的数据库编号
	name *datastruct.RedisObj
	argc int                    //命令数量
	argv []*datastruct.RedisObj //命令值

	cmd     *RedisCommand //当前执行的命令
	lastCmd *RedisCommand //最后执行的命令

	reqType  int
	queryBus *datastruct.SDS // 从客户端的请求获取到的数据

	buf     []byte //准备发回客户端的数据
	bufpos  int    //发回给哭互动那的数据pos
	sentLen int    //已发送的字节数
}
type RedisCommand struct {
	name             datastruct.SDS            //命令名字
	redisCommandFunc func(client *RedisClient) //命令解析方法
}
