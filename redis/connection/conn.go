package connection

import (
	"miniRedis/lib/logger"
	"miniRedis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

/*
	客户端和Redis服务器的连接
*/

const (
	// flagSlave 表示这是一个从与服务器的连接
	flagSlave = uint64(1 << iota)
	// flagSlave 表示这是一个从主服务器的连接
	flagMaster
	// flagMulti 表示这正在一个事务中
	flagMulti
)

// Connection 代表着一个客户端的连接
type Connection struct {
	// 与客户端的网络连接
	conn net.Conn

	// wait until finish sending data, used for graceful shutdown
	// 用于等待完成发送数据，用于优雅关闭。
	sendingData wait.Wait

	// lock while server sending response
	// 用于在服务器发送响应时锁定。
	mu    sync.Mutex
	flags uint64

	// subscribing channels
	// 代表订阅的频道。
	subs map[string]bool

	// password may be changed by CONFIG command during runtime, so store the password
	// 代表客户端的密码，可以在运行时通过 CONFIG 命令更改。
	password string

	// queued commands for `multi`
	// 代表 multi 命令的排队命令。
	queue [][][]byte
	//代表正在观察的键。
	watching map[string]uint32
	// 代表事务中的错误。
	txErrors []error

	// selected db
	// 代表选择的数据库，从0-15
	selectedDB int
}

/*
创建一个连接池（？不推荐使用Pool创建连接池？）
*/
var connPool = sync.Pool{
	New: func() interface{} {
		return &Connection{}
	},
}

// RemoteAddr 返回远程地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 将连接的内容初始化并且放回连接池中
func (c *Connection) Close() error {
	// 等待之前的消息发送完毕
	c.sendingData.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	c.subs = nil
	c.password = ""
	c.queue = nil
	c.watching = nil
	c.txErrors = nil
	c.selectedDB = 0
	connPool.Put(c)
	return nil
}

// NewConn 创建一个客户端连接对象，暂时只保存了net.conn网络连接
func NewConn(conn net.Conn) *Connection {
	// 从池中获取一个对象，减少开支
	c, ok := connPool.Get().(*Connection)
	if !ok {
		logger.Error("connection pool make wrong type")
		return &Connection{
			conn: conn,
		}
	}
	c.conn = conn
	return c
}

// Write 返回消息给TCP连接
func (c *Connection) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	c.sendingData.Add(1)
	defer func() {
		c.sendingData.Done()
	}()

	return c.conn.Write(b)
}

// Name 返回远程地址的地址信息 "192.0.2.1:25"
func (c *Connection) Name() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}

// Subscribe 将当前连接加入到指定的信道的订阅人中
func (c *Connection) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.subs == nil {
		c.subs = make(map[string]bool)
	}
	c.subs[channel] = true
}

func (c *Connection) UnSubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.subs) == 0 {
		return
	}
	delete(c.subs, channel)
}

func (c *Connection) SubsCount() int {
	return len(c.subs)
}

// GetChannels 返回所有正在订阅的信道
func (c *Connection) GetChannels() []string {
	if c.subs == nil {
		return make([]string, 0)
	}
	channels := make([]string, len(c.subs))
	i := 0
	for channel := range c.subs {
		channels[i] = channel
		i++
	}
	return channels
}

func (c *Connection) SetPassword(password string) {
	c.password = password
}

func (c *Connection) GetPassword() string {
	return c.password
}

// InMultiState 判断当前连接是否在事务中
func (c *Connection) InMultiState() bool {
	return c.flags&flagMulti > 0
}

func (c *Connection) SetMultiState(state bool) {
	if !state { // 传入为false，清除事务标志
		c.watching = nil
		c.queue = nil
		c.flags &= ^flagMulti // clean multi flag
		return
	}
	c.flags |= flagMulti
}

// GetQueuedCmdLine 获取事务队列中的命令，一个命令用[][]byte表示，多条命令则[][][]byte
func (c *Connection) GetQueuedCmdLine() [][][]byte {
	return c.queue
}

// EnqueueCmd  将命令加入事务队列
func (c *Connection) EnqueueCmd(cmdLine [][]byte) {
	c.queue = append(c.queue, cmdLine)
}

func (c *Connection) ClearQueuedCmds() {
	c.queue = nil
}

// GetTxErrors 获取事务中的错误信息
func (c *Connection) GetTxErrors() []error {
	return c.txErrors
}

func (c *Connection) AddTxError(err error) {
	c.txErrors = append(c.txErrors, err)
}

// GetWatching 返回正在watch的key，以及其版本编号
func (c *Connection) GetWatching() map[string]uint32 {
	if c.watching == nil {
		c.watching = make(map[string]uint32)
	}
	return c.watching
}

func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}

func (c *Connection) SetSlave() {
	c.flags |= flagSlave
}

func (c *Connection) IsSlave() bool {
	return c.flags&flagSlave > 0
}

func (c *Connection) SetMaster() {
	c.flags |= flagMaster
}

func (c *Connection) IsMaster() bool {
	return c.flags&flagMaster > 0
}
