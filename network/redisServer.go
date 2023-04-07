package network

import (
	"miniRedis/datastruct"
	"miniRedis/db"
	"os"
)

type RedisServer struct {
	pid int //pid

	hz       int              //hz
	db       *db.RedisDb      //db
	commands *datastruct.Dict //redis命令字典，key = sds(命令，比如get/set)， value = *redisCommand

	clientCounter int              //存储client的id计数器
	clients       *datastruct.Dict //客户端字典， key = id, value = *redisClient
	port          int              //端口
	tcpBacklog    int
	bindaddr      string //地址
	ipfdCount     int

	events *EventLoop //事件处理器

	lruclock uint64

	//limits
	maxClients       uint   //max number of simultaneous clients
	maxMemory        uint64 //max number of memory bytes to use
	maxMemoryPolicy  int    //policy for key eviction
	maxMemorySamples int
}

func InitRedisServer() {
	create := datastruct.DictCreate(datastruct.MyDictType{}, nil)
	event := &EventLoop{react: ReactHandler, onOpen: AcceptHandler}
	db := db.DBCreate(0)
	server = &RedisServer{
		pid:     os.Getgid(),
		clients: create,
		events:  event,
		db:      db,
	}
}
