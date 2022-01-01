# CancelChain⛓️💥
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://go.dev/)
[![Documentation](https://godoc.org/github.com/mind1949/cancelchain?status.svg)](https://pkg.go.dev/github.com/mind1949/cancelchain)
[![Go Report Card](https://goreportcard.com/badge/github.com/mind1949/cancelchain)](https://goreportcard.com/report/github.com/mind1949/cancelchain)

CancelChain 提供并发原语，轻松完成顺序关闭go协程。

# 使用
```golang
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mind1949/cancelchain"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c := cancelchain.WithContext(ctx)

	for i := 0; i < 10; i++ {
		seq := i

		c.Go(func(ctx context.Context) error {
			<-ctx.Done()
			fmt.Printf("exit goroutine[%d]\n", seq)
			return ctx.Err()
		})
	}

	err := c.Wait()
	fmt.Printf("err: %+v\n", err)
}

```
