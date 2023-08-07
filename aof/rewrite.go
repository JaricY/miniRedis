package aof

import (
	"io"
	"io/ioutil"
	"miniRedis/config"
	"miniRedis/interface/database"
	"miniRedis/lib/logger"
	"miniRedis/lib/utils"
	"miniRedis/redis/protocol"
	"os"
	"strconv"
	"time"
)

func (persister *Persister) newRewriteHandler() *Persister {
	h := &Persister{}
	h.aofFilename = persister.aofFilename
	h.db = persister.tmpDBMaker()
	return h
}

// RewriteCtx holds context of an AOF rewriting procedure
type RewriteCtx struct {
	tmpFile  *os.File // 重写过程中创建的临时文件指针，用于写入重写后的AOF文件
	fileSize int64    // 当前AOF文件的大小，用于限制重写的最大长度
	dbIdx    int      // 重写过程中选择的数据库索引
}

// Rewrite carries out AOF rewrite
func (persister *Persister) Rewrite() error {
	ctx, err := persister.StartRewrite()
	if err != nil {
		return err
	}
	err = persister.DoRewrite(ctx)
	if err != nil {
		return err
	}

	persister.FinishRewrite(ctx)
	return nil
}

// DoRewrite actually rewrite aof file
// makes DoRewrite public for testing only, please use Rewrite instead
func (persister *Persister) DoRewrite(ctx *RewriteCtx) error {
	tmpFile := ctx.tmpFile

	// load aof tmpFile
	tmpAof := persister.newRewriteHandler()
	tmpAof.LoadAof(int(ctx.fileSize))

	// rewrite aof tmpFile
	for i := 0; i < config.Properties.Databases; i++ {
		// select db
		data := protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(i))).ToBytes()
		_, err := tmpFile.Write(data)
		if err != nil {
			return err
		}
		// dump db
		tmpAof.db.ForEach(i, func(key string, entity *database.DataEntity, expiration *time.Time) bool {
			cmd := EntityToCmd(key, entity)
			if cmd != nil {
				_, _ = tmpFile.Write(cmd.ToBytes())
			}
			if expiration != nil {
				cmd := MakeExpireCmd(key, *expiration)
				if cmd != nil {
					_, _ = tmpFile.Write(cmd.ToBytes())
				}
			}
			return true
		})
	}
	return nil
}

// StartRewrite prepares rewrite procedure
func (persister *Persister) StartRewrite() (*RewriteCtx, error) {
	persister.pausingAof.Lock() // pausing aof
	defer persister.pausingAof.Unlock()

	err := persister.aofFile.Sync() // 刷新之前的AOF文件使之确定持久化成功
	if err != nil {
		logger.Warn("fsync failed")
		return nil, err
	}

	// get current aof file size
	fileInfo, _ := os.Stat(persister.aofFilename) // 获取文件信息
	filesize := fileInfo.Size()

	// create tmp file
	file, err := ioutil.TempFile("", "*.aof")
	if err != nil {
		logger.Warn("tmp file create failed")
		return nil, err
	}
	return &RewriteCtx{
		tmpFile:  file,
		fileSize: filesize,
		dbIdx:    persister.currentDB,
	}, nil
}

// FinishRewrite finish rewrite procedure
func (persister *Persister) FinishRewrite(ctx *RewriteCtx) {
	persister.pausingAof.Lock() // 确保重写期间没有其他写操作，保证数据一致性
	defer persister.pausingAof.Unlock()

	tmpFile := ctx.tmpFile

	src, err := os.Open(persister.aofFilename)
	if err != nil {
		logger.Error("open aofFilename failed: " + err.Error())
		return
	}
	defer func() {
		_ = src.Close()
	}()
	_, err = src.Seek(ctx.fileSize, 0) // seek函数用于修改文件指针的位置，在前面复制完成之后，要在结尾加上SELECT语句
	if err != nil {
		logger.Error("seek failed: " + err.Error())
		return
	}

	// 修改为之前未选择的数据库
	data := protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(ctx.dbIdx))).ToBytes()
	_, err = tmpFile.Write(data)
	if err != nil {
		logger.Error("tmp file rewrite failed: " + err.Error())
		return
	}

	// 将原有的AOF文件的剩余内容复制到tmpFIle中
	_, err = io.Copy(tmpFile, src)
	if err != nil {
		logger.Error("copy aof filed failed: " + err.Error())
		return
	}

	tmpFileName := tmpFile.Name()
	_ = tmpFile.Close()
	// 使用临时的文件代替AOF文件
	_ = persister.aofFile.Close()
	_ = os.Rename(tmpFileName, persister.aofFilename)

	// 修改当前的AOF文件并打开以便后续的操作
	aofFile, err := os.OpenFile(persister.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	persister.aofFile = aofFile

	// 确保当前的数据库和AOF文件中的是同一个数据库索引
	data = protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(persister.currentDB))).ToBytes()
	_, err = persister.aofFile.Write(data)
	if err != nil {
		panic(err)
	}
}
