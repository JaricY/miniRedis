package network

import (
	"fmt"
	"github.com/panjf2000/gnet"
)

type EventLoop struct { //事件轮训
	react  func(frame []byte, c gnet.Conn) (out []byte, action gnet.Action)
	onOpen func(c gnet.Conn) (out []byte, action gnet.Action)
	*gnet.EventServer
}

// 接收来自连接的数据并处理。
func (e *EventLoop) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	cmd := string(frame)
	fmt.Println("Received command:", cmd)
	return e.react(frame, c)
}

// 连接打开时调用。
func (e *EventLoop) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Printf("Accepted connection from %s\n", c.RemoteAddr().String())

	return e.onOpen(c)

}

func ElMain() {

	InitRedisServer()
	err := gnet.Serve(server.events, "tcp://:6399", gnet.WithMulticore(true))

	if err != nil {
		fmt.Println("Error starting server:", err.Error())
		return
	}

}
