package network

import (
	"miniRedis/datastruct"
	"miniRedis/db"
	"os"
)

var (
	redisCommandTable = []*RedisCommand{
		{*datastruct.NewSDS("ping"), pingCommand},
	}
)

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
	InitServerCommandDict()
}

func InitServerCommandDict() {
	server.commands = &datastruct.Dict{}
	for _, c := range redisCommandTable {
		server.commands.DictAdd(c.name, c)
	}
}
