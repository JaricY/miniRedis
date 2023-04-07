package network

import (
	"log"
	"miniRedis/datastruct"
)

// 处理命令
func processCommand(client *RedisClient) int {
	client.Cmd = lookupCommand(client.Argv[0].Ptr.(datastruct.SDS))
	client.LastCmd = client.Cmd

	if client.Cmd == nil {

		return redisOk
	}
	return redisOk
}

// 查找命令
func lookupCommand(name datastruct.SDS) *RedisCommand {
	cmd := server.commands.DictFetchValue(name)
	log.Printf("lookup command: %v", cmd)
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

func pingCommand(c *RedisClient) {

}
