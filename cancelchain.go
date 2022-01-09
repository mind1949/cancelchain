// Package cancelchain 轻松实现并发启动、循序取消goroutine
package cancelchain

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// WithContext 返回一个新的Chain，可通过ctx发送取消信号给第一个由Go方法调用的函数fn。
func WithContext(ctx context.Context) Chain {
	ctxChan := make(chan context.Context, 1)
	ctxChan <- ctx
	return &chain{
		ctxChan: ctxChan,
	}
}

// Chain 编排goroutine，实现并发启动、顺序取消goroutine。
type Chain interface {
	// Go 在一个新的goroutine中调用函数fn
	//
	// 当函数fn退出时，会通过先发送取消信号给第一个被Go调用的函数，
	// 待当前goroutine接受到取消信号后，再发送取消信号给下一个在Go函数中调用的函数
	// 从而形成一个顺序取消goroutine的链状结构。
	Go(fn func(ctx context.Context) error)

	// Wait 阻塞直到所有在Go方法中调用的函数退出
	// 返回的error是所有Go方法调用的函数fn中，第一个退出的函数fn的返回值
	Wait() error
}

type chain struct {
	ctxChan chan context.Context
	wg      sync.WaitGroup

	num int64

	err     error
	onceErr sync.Once

	cancel context.CancelFunc
}

// Go 在一个新的goroutine中调用函数fn
//
// 当函数fn退出时，会通过Context先发送取消信号给第一个被Go调用的函数，
// 待当前goroutine接受到取消信号后，再发送取消信号给下一个在Go函数中调用的函数。
// 从而形成一个顺序取消goroutine的链状结构。
func (c *chain) Go(fn func(ctx context.Context) error) {
	id := atomic.LoadInt64(&c.num)
	atomic.AddInt64(&c.num, 1)

	c.wg.Add(1)

	currentCtx := <-c.ctxChan
	if id == 0 {
		currentCtx, c.cancel = context.WithCancel(currentCtx)
	}
	nextCtx, cancel := context.WithCancel(context.Background())
	c.ctxChan <- nextCtx

	go func() {
		defer func() {
			<-currentCtx.Done()
			cancel()
			c.wg.Done()
		}()

		err := fn(currentCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			c.cancel()
			c.onceErr.Do(func() {
				c.err = err
			})
		}
		if id == 0 {
			c.cancel()
		}
	}()
}

// Wait 阻塞直到所有在Go方法中调用的函数退出
//
// 返回的error是所有Go方法调用的函数fn中，第一个退出的函数fn的返回值
func (c *chain) Wait() error {
	c.wg.Wait()
	return c.err
}
