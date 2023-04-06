package tcp

import (
	"fmt"
	"log"
	"net"
)

func ListenAndServe(address string) {
	//开始监听对应的地址
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("listen err: %v", err))
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			// 通常是由于listener被关闭无法继续监听导致的错误
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}
		//开启一个新的协程处理该连接
		go Handle(conn)
	}
}

func Handle(conn net.Conn) {

}
