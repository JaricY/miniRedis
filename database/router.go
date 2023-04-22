package database

import "strings"

var cmdTable = make(map[string]*command)

// 用于实现一个Redis的命令解析和执行，每一个命令都对应一个command结构体
type command struct {
	executor ExecFunc // 执行的函数
	prepare  PreFunc  // 用于准备相关命令操作的函数（例如加锁）
	undo     UndoFunc // 撤销命令的函数
	arity    int      // 表示命令所需参数的数量，允许负数，负数表示参数的数量至少为该值的绝对值
	flags    int      // 表示命令的标志，用于标识命令的属性，例如是否支持事务、是否支持读写等
}

const (
	flagWrite    = 0
	flagReadOnly = 1
)

// RegisterCommand registers a new command
// arity means allowed number of cmdArgs, arity < 0 means len(args) >= -arity.
// for example: the arity of `get` is 2, `mget` is -2
func RegisterCommand(name string, executor ExecFunc, prepare PreFunc, rollback UndoFunc, arity int, flags int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		prepare:  prepare,
		undo:     rollback,
		arity:    arity,
		flags:    flags,
	}
}

func isReadOnlyCommand(name string) bool {
	name = strings.ToLower(name)
	cmd := cmdTable[name]
	if cmd == nil {
		return false
	}
	return cmd.flags&flagReadOnly > 0
}
