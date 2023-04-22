package protocol

import (
	"bytes"
	"miniRedis/interface/redis"
	"strconv"
)

var (
	// CRLF 是RedisRESP协议中每行的末尾分隔符
	CRLF = "\r\n"
)

// BulkReply 存储一行二进制安全的字符串
type BulkReply struct {
	Arg []byte
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

// ToBytes 返回用于响应客户端的单行数组，例如$5\r\nvalue\r\n
func (r *BulkReply) ToBytes() []byte {
	if r.Arg == nil {
		return nullBulkBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

// MultiBulkReply 存储的是数组类型，返回多个（切片）字符串类型（byte数组类型）的响应结果
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply 创建一个多行字符串的返回结果
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

// ToBytes 返回用于响应客户端的多行数组，例如：
// *1\r\n
// $4\r\n
// ping\r\n
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

// 返回的是多行回复
type MultiRawReply struct {
	Replies []redis.Reply
}

func MakeMultiRawReply(replies []redis.Reply) *MultiRawReply {
	return &MultiRawReply{
		Replies: replies,
	}
}

// ToBytes 多行回复，等于将每行都执行对应的toByte之后再拼接
func (r *MultiRawReply) ToBytes() []byte {
	argLen := len(r.Replies)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Replies {
		buf.Write(arg.ToBytes())
	}
	return buf.Bytes()
}

// StatusReply 简单状态回复
type StatusReply struct {
	Status string
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

func IsOKReply(reply redis.Reply) bool {
	return string(reply.ToBytes()) == "+OK\r\n"
}

type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

type StandardErrReply struct {
	Status string
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

func IsErrorReply(reply redis.Reply) bool {
	return reply.ToBytes()[0] == '-'
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}
