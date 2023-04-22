package redis

// Connection 是表示客户端和服务端之间的一个连接，接口类型定义了需要连接的方法
type Connection interface {
	Write([]byte) (int, error)
	Close() error

	SetPassword(string)
	GetPassword() string
	// client should keep its subscribing channels
	Subscribe(channel string)
	UnSubscribe(channel string)
	SubsCount() int
	GetChannels() []string

	InMultiState() bool
	SetMultiState(bool)
	GetQueuedCmdLine() [][][]byte
	EnqueueCmd([][]byte)
	ClearQueuedCmds()
	GetWatching() map[string]uint32
	AddTxError(err error)
	GetTxErrors() []error

	GetDBIndex() int
	SelectDB(int)

	SetSlave()
	IsSlave() bool

	SetMaster()
	IsMaster() bool

	Name() string
}
