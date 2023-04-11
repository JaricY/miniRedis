package tcp

import (
	"context"
	"fmt"
	"miniRedis/interface/tcp"
	"miniRedis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Config 保存了创建TCP连接的配置信息
type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

// ClientCounter 用于记录连接到miniRedis的客户端数量
var ClientCounter int

// ListenAndServeWithSignal 用于监听和处理请求，并且携带信号量用于处理异常，例如请求关闭等情况
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	// 创建一个管道，记录请求关闭信号
	closeChan := make(chan struct{})
	// 创建一个管道，接受操作系统发送的信号
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	// 开启一个新的协程等待操作系统的信号
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	//开始监听，返回一个TCP监听器
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe 绑定端口并处理请求，持续阻塞直到关闭
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	errCh := make(chan error, 1)
	defer close(errCh)
	// 开启一个协程处理关闭和错误信息
	go func() {
		select { // 阻塞接受
		case <-closeChan:
			logger.Info("get exit signal")
		case er := <-errCh:
			logger.Info(fmt.Sprintf("accept error: %s", er.Error()))
		}
		logger.Info("shutting down...")
		_ = listener.Close() // 关闭监听器
		_ = handler.Close()  // 关闭连接
	}()

	ctx := context.Background() // 获取请求上下文

	var waitDone sync.WaitGroup // 用于等待所有的协程执行结束后才执行 listener.Close() 和 handler.Close()
	for {

		// 如果接受错误则写入到管道中
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			break
		}

		logger.Info("accept link")
		ClientCounter++
		waitDone.Add(1)
		//异步执行
		go func() {
			defer func() {
				waitDone.Done()
				ClientCounter--
			}()
			//handle是对整个连接的
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
