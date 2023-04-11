package database

import (
	"miniRedis/datastruct/dict"
	"miniRedis/datastruct/lock"
)

const (
	dataDictSize = 1 << 16
	ttlDictSize  = 1 << 10
	lockerSize   = 1024
)

type DB struct {
	index int
	// key -> DataEntity
	data dict.Dict
	// key -> expireTime (time.Time)
	ttlMap dict.Dict
	// key -> version(uint32)
	versionMap dict.Dict

	// dict.Dict will ensure concurrent-safety of its method
	// use this mutex for complicated command only, eg. rpush, incr ...
	locker *lock.Locks
	addAof func(CmdLine)
}

// CmdLine 一个CmdLIne表示一个命令行，因为命令行有行，所以使用二维数组
type CmdLine = [][]byte
