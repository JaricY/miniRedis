package network

import "github.com/panjf2000/gnet"

type EventLoop struct { //事件轮训
	react  func(frame []byte, c gnet.Conn) (out []byte, action gnet.Action)
	accept func(c gnet.Conn) (out []byte, action gnet.Action)

	*gnet.EventServer
}

// 读事件处理
func (e *EventLoop) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	return e.react(frame, c)
}

// 新连接处理
func (e *EventLoop) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return e.accept(c)
}

func elMain() {

}
