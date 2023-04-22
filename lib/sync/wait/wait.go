package wait

import (
	"sync"
	"time"
)

/*
	在sync.WaitGroup的基础上加了超时等待，本质上仍然是一个sync.WaitGroup
*/

// Wait is similar with sync.WaitGroup which can wait with timeout
type Wait struct {
	wg sync.WaitGroup
}

// Add adds delta, which may be negative, to the WaitGroup counter.
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// Done decrements the WaitGroup counter by one
func (w *Wait) Done() {
	w.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero.
func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout 将会阻塞等待，直到调用 w.Signal() 或者 w.Broadcast() 方法通知结束。
// 或者在timeout时间后也会表示超时，如果超时则直接执行后序内容
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		w.Wait()
		c <- struct{}{}
	}()
	select {
	case <-c:
		return false // 正常结束
	case <-time.After(timeout):
		return true // 超时结束
	}
}
