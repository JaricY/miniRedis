package redis

// Reply 是RESP协议的返回消息内容
type Reply interface {
	ToBytes() []byte
}
