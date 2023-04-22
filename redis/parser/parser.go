package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"miniRedis/interface/redis"
	"miniRedis/lib/logger"
	"miniRedis/redis/protocol"
	"runtime/debug"
	"strconv"
)

// 解析客户端发送的内容

// Payload 存储了redis服务器解析得到的回复或者错误
type Payload struct {
	Data redis.Reply
	Err  error
}

// 解析客户端请求
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	//使用管道实现并行处理
	go parse0(reader, ch)
	return ch
}

// ParseBytes reads data from []byte and return all replies
func ParseBytes(data []byte) ([]redis.Reply, error) {
	ch := make(chan *Payload)
	reader := bytes.NewReader(data)
	go parse0(reader, ch)
	var results []redis.Reply
	for payload := range ch {
		if payload == nil {
			return nil, errors.New("no protocol")
		}
		if payload.Err != nil {
			if payload.Err == io.EOF {
				break
			}
			return nil, payload.Err
		}
		results = append(results, payload.Data)
	}
	return results, nil
}

// ParseOne 读取一行数据并返回第一个解析到的内容
func ParseOne(data []byte) (redis.Reply, error) {
	ch := make(chan *Payload)
	reader := bytes.NewReader(data)
	go parse0(reader, ch)
	payload := <-ch // parse0 will close the channel
	if payload == nil {
		return nil, errors.New("no protocol")
	}
	return payload.Data, payload.Err
}

// 解析每行的内容
func parse0(rawReader io.Reader, ch chan<- *Payload) {
	// 发生异常，记录
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err, string(debug.Stack()))
		}
	}()

	//net.conn 只存储客户端发送给服务端的数据
	reader := bufio.NewReader(rawReader)
	for {
		// 读取出一行,以\n结尾
		line, err := reader.ReadBytes('\n')

		//当读取错误的时候说明已经读取完成
		if err != nil {
			ch <- &Payload{Err: err}
			close(ch)
			return
		}
		length := len(line)

		// 说明是空行，不处理
		if length <= 2 || line[length-2] != '\r' {
			// there are some empty lines within replication traffic, ignore this error
			//protocolError(ch, "empty line")
			continue
		}
		// 从尾部删除\r\n之后的切片
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] { // 读取协议类型
		case '+':

			content := string(line[1:])
			ch <- &Payload{
				// 返回一个状态的回复
				Data: protocol.MakeStatusReply(content),
			}
		case '-':
			//说明是一个错误
			ch <- &Payload{
				Data: protocol.MakeErrReply(string(line[1:])),
			}
		case ':':
			// 说明是一个整数
			value, err := strconv.ParseInt(string(line[1:]), 10, 64)
			if err != nil {
				protocolError(ch, "illegal number "+string(line[1:]))
				continue
			}
			ch <- &Payload{
				Data: protocol.MakeIntReply(value),
			}
		case '$':
			// 单行的字符串，二进制安全的
			err = parseBulkString(line, reader, ch)
			if err != nil {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
		case '*':
			// 数组类型 line的当前行，剩下的交给parseArray处理，reader是得到的全部数据，ch写入内容的管道，
			err = parseArray(line, reader, ch)
			if err != nil {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
		default:
			args := bytes.Split(line, []byte{' '})
			ch <- &Payload{
				Data: protocol.MakeMultiBulkReply(args),
			}
		}
	}
}

// parseBulkString 解析单行命令
func parseBulkString(header []byte, reader *bufio.Reader, ch chan<- *Payload) error {
	// 当前header只包含 $长度
	strLen, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || strLen < -1 {
		protocolError(ch, "illegal bulk string header: "+string(header))
		return nil
	} else if strLen == -1 {
		ch <- &Payload{
			Data: protocol.MakeNullBulkReply(),
		}
		return nil
	}
	body := make([]byte, strLen+2)
	// 继续获取
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	ch <- &Payload{
		Data: protocol.MakeBulkReply(body[:len(body)-2]),
	}
	return nil
}

// 解析数组类型（*开头）
func parseArray(header []byte, reader *bufio.Reader, ch chan<- *Payload) error {
	/*
		*2\r\n
		 $3\r\n
		 get
		 $2\r\n
		 k1
	*/

	//获取长度
	nStrs, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || nStrs < 0 {
		protocolError(ch, "illegal array header "+string(header[1:]))
		return nil
	} else if nStrs == 0 {
		ch <- &Payload{
			Data: protocol.MakeEmptyMultiBulkReply(),
		}
		return nil
	}

	// nstrs表示共有多少行，lines用于存储命令参数，是一个切片，切片元素是字节数组
	lines := make([][]byte, 0, nStrs)
	// 处理每一行
	for i := int64(0); i < nStrs; i++ {
		var line []byte
		// 处理一行以\n结尾(在上个方法中已经读了一行，所以这里是第二行开始)
		// line = $3\r\n
		line, err = reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		length := len(line)

		//判断是否是一行数据
		if length < 4 || line[length-2] != '\r' || line[0] != '$' {
			protocolError(ch, "illegal bulk string header "+string(line))
			break
		}
		// strlen = 3
		strLen, err := strconv.ParseInt(string(line[1:length-2]), 10, 64)
		if err != nil || strLen < -1 {
			protocolError(ch, "illegal bulk string length "+string(line))
			break
		} else if strLen == -1 {
			lines = append(lines, []byte{})
		} else {
			//strlen+2 = 5
			body := make([]byte, strLen+2)
			//读取一行数据 读取了get
			_, err := io.ReadFull(reader, body)
			if err != nil {
				return err
			}
			lines = append(lines, body[:len(body)-2])
		}
	}
	ch <- &Payload{
		Data: protocol.MakeMultiBulkReply(lines),
	}
	return nil
}

func protocolError(ch chan<- *Payload, msg string) {
	err := errors.New("protocol error: " + msg)
	ch <- &Payload{Err: err}
}
