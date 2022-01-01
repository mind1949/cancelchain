// Package cancelchain 提供顺序取消go协程的同步原语
package cancelchain

import (
	"context"
	"sync"
	"sync/atomic"
)

// WithContext 返回一个新的Chain，通过ctx发送取消信号给第一个由Go方法调用的函数fn。
func WithContext(ctx context.Context) Chain {
	ctxChan := make(chan context.Context, 1)
	ctxChan <- ctx
	return &chain{
		ctxChan: ctxChan,
	}
}

// Chain 编排go协程，在逻辑上形成顺序取消go协程的链状结构。
type Chain interface {
	// Go 在一个新的go协程中调用函数fn
	// 函数fn退出会通过Context发送取消信号给下一个在Go函数中调用的函数fn
	// 在逻辑上形成一个顺序取消go协程的链状结构。
	Go(fn func(ctx context.Context) error)
	// Wait 阻塞直到所有在Go方法中调用的函数退出
	// 返回的error是第一个在Go方法中调用的函数fn的返回值
	Wait() error
}

type chain struct {
	ctxChan chan context.Context
	wg      sync.WaitGroup

	num int64
	err error
}

// Go 在一个新的go协程中调用函数fn
//
// 当函数fn退出时，会通过Context发送取消信号给下一个在Go函数中调用的函数fn，
// 从而形成一个顺序取消go协程的链状结构。
func (c *chain) Go(fn func(ctx context.Context) error) {
	id := atomic.LoadInt64(&c.num)
	atomic.AddInt64(&c.num, 1)

	c.wg.Add(1)

	currentCtx := <-c.ctxChan
	nextCtx, cancel := context.WithCancel(context.Background())
	c.ctxChan <- nextCtx

	go func() {
		err := fn(currentCtx)
		defer func() {
			cancel()
			c.wg.Done()
		}()
		if id == 0 {
			c.err = err
		}
	}()
}

// Wait 阻塞直到所有在Go方法中调用的函数退出
// 返回的error是第一个在Go方法中调用的函数fn的返回值
func (c *chain) Wait() error {
	c.wg.Wait()
	return c.err
}
