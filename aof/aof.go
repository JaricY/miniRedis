package aof

import (
	"context"
	"io"
	"miniRedis/interface/database"
	"miniRedis/lib/logger"
	"miniRedis/lib/utils"
	"miniRedis/redis/connection"
	"miniRedis/redis/parser"
	"miniRedis/redis/protocol"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16
)

const (
	// FsyncAlways do fsync for every command
	FsyncAlways = "always"
	// FsyncEverySec do fsync every second
	FsyncEverySec = "everysec"
	// FsyncNo lets operating system decides when to do fsync
	FsyncNo = "no"
)

type payload struct { // 用于存储写入AOF文件的相关信息
	cmdLine CmdLine         // 用于存储一个命令行
	dbIndex int             // 对应的数据库下标
	wg      *sync.WaitGroup // 用于等待多个goroutine完成AOF文件操作
}

type Listener interface {
	// Callback will be called-back after receiving a aof payload
	Callback([]CmdLine)
}

// Persister receive msgs from channel and write to AOF file
type Persister struct { // 用于执行持久化的执行者
	ctx         context.Context          //持久化协程的上下文
	cancel      context.CancelFunc       // 持久化协程的取消函数
	db          database.DBEngine        // 数据库引擎
	tmpDBMaker  func() database.DBEngine // 临时的数据库创建函数，用于AOF重写的时候创建临时数据库
	aofChan     chan *payload            // 用于在持久化协程和Redis主协程之间传递任务的通道，一般用于在AOF重写的时候作为临时的重写缓冲区
	aofFile     *os.File                 // AOF文件。
	aofFilename string                   // AOF名。
	aofFsync    string                   // AOF写入策略（always/everysec/no）
	// aof goroutine will send msg to main goroutine through this channel when aof tasks finished and ready to shut down
	// 当aof任务完成并准备关闭时，aof goroutine将通过此通道向main goroutine发送消息。
	aofFinished chan struct{} // 持久化协程完成后通知Redis主协程的通道
	// pause aof for start/finish aof rewrite progress
	pausingAof sync.Mutex            // 锁，用于在AOF重写期间暂停AOF持久化
	currentDB  int                   // 当前正在使用的数据库编号
	listeners  map[Listener]struct{} // Redis事件监听器
	// reuse cmdLine buffer
	buffer []CmdLine // 命令缓冲区
}

// NewPersister creates a new aof.Persister
func NewPersister(db database.DBEngine, filename string, load bool, fsync string, tmpDBMaker func() database.DBEngine) (*Persister, error) {
	persister := &Persister{}
	persister.aofFilename = filename
	persister.aofFsync = strings.ToLower(fsync)
	persister.db = db
	persister.tmpDBMaker = tmpDBMaker
	persister.currentDB = 0
	if load {
		persister.LoadAof(0)
	}
	aofFile, err := os.OpenFile(persister.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	persister.aofFile = aofFile
	persister.aofChan = make(chan *payload, aofQueueSize)
	persister.aofFinished = make(chan struct{})
	persister.listeners = make(map[Listener]struct{})
	go func() {
		persister.listenCmd()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	persister.ctx = ctx
	persister.cancel = cancel
	if persister.aofFsync == FsyncEverySec {
		persister.fsyncEverySecond()
	}
	return persister, nil
}

func (persister *Persister) RemoveListener(listener Listener) {
	persister.pausingAof.Lock()
	defer persister.pausingAof.Unlock()
	delete(persister.listeners, listener)
}

func (persister *Persister) SaveCmdLine(dbIndex int, cmdLine CmdLine) {
	// aofChan will be set as nil temporarily during load aof see Persister.LoadAof
	if persister.aofChan == nil {
		return
	}
	// FsyncAlways 策略表示每个命令都要进行AOF操作
	if persister.aofFsync == FsyncAlways {
		p := &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
		persister.writeAof(p)
		return
	}
	persister.aofChan <- &payload{
		cmdLine: cmdLine,
		dbIndex: dbIndex,
	}
}

func (persister *Persister) listenCmd() {
	for p := range persister.aofChan {
		persister.writeAof(p)
	}
	persister.aofFinished <- struct{}{}
}

func (persister *Persister) writeAof(p *payload) {
	persister.buffer = persister.buffer[:0] // 清空缓冲区以便后续复用
	persister.pausingAof.Lock()             // prevent other goroutines from pausing aof
	defer persister.pausingAof.Unlock()
	// ensure aof is in the right database
	if p.dbIndex != persister.currentDB {
		// 修改数据库
		selectCmd := utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))
		persister.buffer = append(persister.buffer, selectCmd)
		data := protocol.MakeMultiBulkReply(selectCmd).ToBytes()
		_, err := persister.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
			return // skip this command
		}
		persister.currentDB = p.dbIndex
	}

	// save command
	data := protocol.MakeMultiBulkReply(p.cmdLine).ToBytes()
	persister.buffer = append(persister.buffer, p.cmdLine)
	_, err := persister.aofFile.Write(data)
	if err != nil {
		logger.Warn(err)
	}
	for listener := range persister.listeners {
		listener.Callback(persister.buffer)
	}
	if persister.aofFsync == FsyncAlways {
		_ = persister.aofFile.Sync() // 同步刷新到磁盘
	}
}

// 将aof文件加载到内存
func (persister *Persister) LoadAof(maxBytes int) {
	// persister.db.Exec may call persister.addAof
	// delete aofChan to prevent loaded commands back into aofChan
	aofChan := persister.aofChan // 备份aofChan
	persister.aofChan = nil
	defer func(aofChan chan *payload) {
		persister.aofChan = aofChan // 恢复aofchan
	}(aofChan)

	file, err := os.Open(persister.aofFilename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return
		}
		logger.Warn(err)
		return
	}
	defer file.Close()

	var reader io.Reader
	if maxBytes > 0 { // 如果大于0说明只需要读取一部分
		reader = io.LimitReader(file, int64(maxBytes))
	} else {
		reader = file
	}

	// 返回的ch是Payload类型，存储服务器解析到的数据
	ch := parser.ParseStream(reader)
	fakeConn := connection.NewFakeConn() // 创建一个虚拟连接，只用于保存当前的 dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF { // 如果是读到了文件末尾则退出
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil { // 读到了空命令
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*protocol.MultiBulkReply) // 解析到MultiBulkReply
		if !ok {
			logger.Error("require multi bulk protocol")
			continue
		}
		ret := persister.db.Exec(fakeConn, r.Args)
		if protocol.IsErrorReply(ret) {
			logger.Error("exec err", string(ret.ToBytes()))
		}
		if strings.ToLower(string(r.Args[0])) == "select" {
			// execSelect success, here must be no error
			dbIndex, err := strconv.Atoi(string(r.Args[1]))
			if err == nil {
				persister.currentDB = dbIndex
			}
		}
	}
}

func (persister *Persister) Close() {
	if persister.aofFile != nil {
		close(persister.aofChan)
		<-persister.aofFinished // wait for aof finished
		err := persister.aofFile.Close()
		if err != nil {
			logger.Warn(err)
		}
	}
	persister.cancel()
}
func (persister *Persister) fsyncEverySecond() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				persister.pausingAof.Lock()
				if err := persister.aofFile.Sync(); err != nil {
					logger.Errorf("fsync failed: %v", err)
				}
				persister.pausingAof.Unlock()
			case <-persister.ctx.Done():
				return
			}
		}
	}()
}
