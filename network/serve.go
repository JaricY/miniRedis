package network

import (
	"github.com/panjf2000/gnet"
	"miniRedis/datastruct"
	"miniRedis/db"
)

type RedisServer struct { // 服务端的结构
	pid     int //pid
	db      *db.RedisDb
	command *datastruct.Dict // 命令字典，key = sds ， value =

	//存储客户端
	clientCounter int              //客户端的数量
	clients       *datastruct.Dict //客户端字典

	port       int    //端口
	bindaddr   string //地址
	tcpBacklog int    //最大连接数
	ipfdCount  int    //正在监听的套接字文件描述符（fd）的数量
}
type RedisClient struct {
	id   int
	conn gnet.Conn //客户端连接
	db   *db.RedisDb
}
