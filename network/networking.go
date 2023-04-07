package network

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"log"
	"miniRedis/datastruct"
	"runtime"
	"strconv"
	"strings"
)

const (
	redisOk  = 0
	redisErr = -1
)

//网络处理模块，负责客户端和服务端的网络通信

var server = &RedisServer{}

// ReactHandler 处理接收到的请求
func ReactHandler(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	client := server.clients.DictFetchValue(c.Context()).(*RedisClient) //从字典中获取到client客户端对象

	client.QueryBus = datastruct.NewSDS(string(frame)) // 创建查询的sds

	defer func() { //处理异常结果
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Print(err)
			log.Printf("==> %s \n", string(buf[:n]))
		}
	}()
	//client.ProcessInputBuffer() //处理请求
	//out = []byte("+PONG\r\n")
	fmt.Println("frame:", frame)
	fmt.Println("queryBus:", client.QueryBus)
	c.AsyncWrite([]byte("+PONG\r\n"))
	return out, action
}

// AcceptHandler 用于处理接受到新连接并创建一个客户端对象
func AcceptHandler(c gnet.Conn) (out []byte, action gnet.Action) {
	client := CreateClient(c)
	server.clients.DictAdd(client.Id, client)
	return out, action
}

// ProcessInputBuffer 处理客户端接收到的数据
func (c *RedisClient) ProcessInputBuffer() {
	if c.QueryBus.GetIndex(0) == ' ' {
		return
	} else if c.QueryBus.GetIndex(0) == '*' {
		c.ReqType = redisReqMultibulk
	} else {
		c.ReqType = redisReqInline
	}
	log.Printf("client reqtype is : %v", c.ReqType)

	//进行协议的解析
	if c.ReqType == redisReqMultibulk {
		if processMultiBulkBuffer(c) == redisErr {
			//error
			log.Printf("analysis protocol error")
		}
	} else if c.ReqType == redisReqInline {
		if processInlineBuffer(c) == redisErr {
			//error
		}
	}

	if c.Argc == 0 {
		//如果参数的数量为0
		//TODO
	} else {
		processCommand(c)
	}
}

// 解析以 "*" 开头的数据。
func processMultiBulkBuffer(client *RedisClient) int {
	// 按行划分
	newLines := strings.Split(client.QueryBus.String(), "\n")
	//处理每行的参数
	argIdx := 0
	for i, line := range newLines {
		// 处理每行的结尾
		line = strings.Replace(line, "\r", "", 1)
		if i == 0 {
			//arg count
			var err error
			client.Argc, err = strconv.Atoi(line[1:])
			if err != nil {
				return redisErr
			}
			client.Argv = make([]*datastruct.RedisObj, client.Argc)
			continue
		}
		if client.Argc <= argIdx {
			break
		}
		if line[0] != '$' {
			// 不是以'$'符结尾就暂且认为是字符串,也就是命令参数
			client.Argv[argIdx] = datastruct.CreateObject(datastruct.RedisTypeString, datastruct.NewSDS(line))
			argIdx++
		}
	}
	log.Printf("analysis command, command count: %v, value: %v", client.Argc, client.Argv)
	return redisOk
}

// 解析以 "+" 或 "-" 开头的字符串类型
func processInlineBuffer(client *RedisClient) int {
	return 0
}

// 解析以 "$" 开头的数据
func processBulkBuffer(client *RedisClient) int {
	return 0
}

func GenerateClientId() int {
	server.clientCounter++
	return server.clientCounter
}
