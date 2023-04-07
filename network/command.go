package network

import (
	"log"
	"minRedis/datastruct"
)

// 处理命令
func processCommand(client *RedisClient) int {
	log.Printf("开始执行命令")
	client.Cmd = lookupCommand(client.Argv[0].Ptr.(datastruct.SDS))
	client.LastCmd = client.Cmd

	if client.Cmd == nil {

		return redisOk
	}

	call(client)

	return redisOk
}

// 查找命令
func lookupCommand(name datastruct.SDS) *RedisCommand {
	cmd := server.commands.DictFetchValue(name)
	if cmd == nil {
		return nil
	}
	return cmd.(*RedisCommand)
}

func addReplyToBuffer(client *RedisClient, robj *datastruct.RedisObj) {
	sds := robj.Ptr.(datastruct.SDS)
	str := sds.String()
	copy(client.Buf[client.Bufpos:], str)
	client.Bufpos = client.Bufpos + len([]byte(str))
}

func sendReplyToClient(client *RedisClient) int {
	return redisOk
}

func call(client *RedisClient) {
	log.Printf("执行call（）")
	client.Cmd.redisCommandFunc(client)
}

func pingCommand(c *RedisClient) {
	c.Conn.AsyncWrite([]byte("+PONG\r\n"))
	log.Printf("成功写入PONG")
}

func infoCommand(c *RedisClient) {
	c.Conn.AsyncWrite([]byte("+PONG\r\n"))
	log.Printf("成功写入PONG")
}
