package network

import (
	"log"
	"miniRedis/datastruct"
)

// 处理命令
func processCommand(client *RedisClient) {
	client.Cmd = lookupCommand(client.Argv[0].Ptr.(datastruct.SDS))

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

func pingCommand(c *RedisClient) {

}
