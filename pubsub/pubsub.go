package pubsub

import (
	"miniRedis/datastruct/list"
	"miniRedis/interface/redis"
	"miniRedis/lib/utils"
	"miniRedis/redis/protocol"
	"strconv"
)

var (
	_subscribe         = "subscribe"
	_unsubscribe       = "unsubscribe"
	messageBytes       = []byte("message")
	unSubscribeNothing = []byte("*3\r\n$11\r\nunsubscribe\r\n$-1\n:0\r\n")
)

// 第一个参数是消息类型，第二个参数是频道名称，第三个参数是消息内容的状态码，包括当前订阅的数量等
func makeMsg(t string, channel string, code int64) []byte {
	return []byte("*3\r\n$" + strconv.FormatInt(int64(len(t)), 10) + protocol.CRLF + t + protocol.CRLF +
		"$" + strconv.FormatInt(int64(len(channel)), 10) + protocol.CRLF + channel + protocol.CRLF +
		":" + strconv.FormatInt(code, 10) + protocol.CRLF)
}

// subscribe0 客户端订阅channel，如果是一个新的channel则返回true
func subscribe0(hub *Hub, channel string, client redis.Connection) bool {
	client.Subscribe(channel)

	// 从字典中获取这个channel对应的订阅者
	raw, ok := hub.subs.Get(channel)
	var subscribers *list.LinkedList
	if ok {
		// 如果存在该channel
		subscribers, _ = raw.(*list.LinkedList)
	} else {
		// 如果不存在该channel
		subscribers = list.Make()
		hub.subs.Put(channel, subscribers)
	}
	// 判断得到的订阅者列表是否包含当前订阅者（如果是新的列表肯定不包含）
	if subscribers.Contains(func(a interface{}) bool {
		return a == client
	}) {
		return false
	}

	subscribers.Add(client)
	return true
}

func unsubscribe0(hub *Hub, channel string, client redis.Connection) bool {
	client.UnSubscribe(channel)

	// remove from hub.subs
	raw, ok := hub.subs.Get(channel)
	if ok {
		subscribers, _ := raw.(*list.LinkedList)
		subscribers.RemoveAllByVal(func(a interface{}) bool {
			return utils.Equals(a, client)
		})

		if subscribers.Len() == 0 {
			// clean
			hub.subs.Remove(channel)
		}
		return true
	}
	return false
}

// Subscribe 将客户端加入到指定的通道中
func Subscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {

	channels := make([]string, len(args))
	for i, b := range args {
		channels[i] = string(b)
	}

	// 加锁保证同步性
	hub.subsLocker.Locks(channels...)
	defer hub.subsLocker.UnLocks(channels...)

	for _, channel := range channels {
		if subscribe0(hub, channel, c) {
			_, _ = c.Write(makeMsg(_subscribe, channel, int64(c.SubsCount())))
		}
	}
	return &protocol.NoReply{}
}

// UnsubscribeAll 移除所有订阅的频道
func UnsubscribeAll(hub *Hub, c redis.Connection) {
	channels := c.GetChannels()

	hub.subsLocker.Locks(channels...)
	defer hub.subsLocker.UnLocks(channels...)

	for _, channel := range channels {
		unsubscribe0(hub, channel, c)
	}

}

// UnSubscribe 移除某个订阅的频道
func UnSubscribe(db *Hub, c redis.Connection, args [][]byte) redis.Reply {
	var channels []string
	if len(args) > 0 {
		channels = make([]string, len(args))
		for i, b := range args {
			channels[i] = string(b)
		}
	} else {
		channels = c.GetChannels()
	}

	db.subsLocker.Locks(channels...)
	defer db.subsLocker.UnLocks(channels...)

	if len(channels) == 0 {
		_, _ = c.Write(unSubscribeNothing)
		return &protocol.NoReply{}
	}

	for _, channel := range channels {
		if unsubscribe0(db, channel, c) {
			_, _ = c.Write(makeMsg(_unsubscribe, channel, int64(c.SubsCount())))
		}
	}
	return &protocol.NoReply{}
}

// Publish 客户端发送信息到频道中
func Publish(hub *Hub, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return &protocol.ArgNumErrReply{Cmd: "publish"}
	}
	channel := string(args[0])
	message := args[1]

	hub.subsLocker.Lock(channel)
	defer hub.subsLocker.UnLock(channel)

	// 获取订阅者
	raw, ok := hub.subs.Get(channel)
	if !ok {
		return protocol.MakeIntReply(0)
	}

	subscribers, _ := raw.(*list.LinkedList)
	subscribers.ForEach(func(i int, c interface{}) bool {
		client, _ := c.(redis.Connection)
		replyArgs := make([][]byte, 3)
		replyArgs[0] = messageBytes    // "message"
		replyArgs[1] = []byte(channel) // "ch1"
		replyArgs[2] = message         // "message1"
		_, _ = client.Write(protocol.MakeMultiBulkReply(replyArgs).ToBytes())
		return true
	})

	// 返回的是接收到的订阅者的数量
	return protocol.MakeIntReply(int64(subscribers.Len()))
}
