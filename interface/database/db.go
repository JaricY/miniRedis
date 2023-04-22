package database

import (
	"miniRedis/interface/redis"
	"time"
)

type CmdLine = [][]byte

type DB interface {
	Exec(client redis.Connection, cmdLine [][]byte) redis.Reply
	AfterClientClose(c redis.Connection)
	Close()
}

type DBEngine interface {
	DB
	ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply
	ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []CmdLine) redis.Reply
	GetUndoLogs(dbIndex int, cmdLine [][]byte) []CmdLine
	ForEach(dbIndex int, cb func(key string, data *DataEntity, expiration *time.Time) bool)
	RWLocks(dbIndex int, writeKeys []string, readKeys []string)
	RWUnLocks(dbIndex int, writeKeys []string, readKeys []string)
	GetDBSize(dbIndex int) (int, int)
}

// DataEntity 存储key的内容，包括string,list,hash等
type DataEntity struct {
	Data interface{}
}
