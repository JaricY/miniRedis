package network

import (
	"minRedis/datastruct"
	"minRedis/db"
	"os"
)

var (
	redisCommandTable = []*RedisCommand{
		{*datastruct.NewSDS("PING"), pingCommand},
		{*datastruct.NewSDS("INFO"), infoCommand},
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
	server.commands = datastruct.DictCreate(datastruct.MyDictType{}, nil)
	for _, c := range redisCommandTable {
		server.commands.DictAdd(c.name, c)
	}
}
