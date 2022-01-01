package cancelchain

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ids := make([]int, 0, 10)
	c := WithContext(ctx)
	for i := 0; i < 10; i++ {
		id := i
		c.Go(func(ctx context.Context) error {
			<-ctx.Done()
			ids = append(ids, id)
			return ctx.Err()
		})
	}
	c.Wait()
	for i := range ids {
		if ids[i] != i {
			t.Fatalf("expect %d got %d", i, ids[i])
		}
	}
}

func TestWithContext(t *testing.T) {
	cases := []struct {
		ctxFn func() (context.Context, func())
		err   error
	}{
		{
			ctxFn: func() (context.Context, func()) {
				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(time.Second, cancel)
				return ctx, cancel
			},
			err: context.Canceled,
		},
		{
			ctxFn: func() (context.Context, func()) {
				return context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
			},
			err: context.DeadlineExceeded,
		},
		{
			ctxFn: func() (context.Context, func()) {
				return context.WithTimeout(context.Background(), 1*time.Second)
			},
			err: context.DeadlineExceeded,
		},
	}

	for _, tc := range cases {
		ctx, cancel := tc.ctxFn()
		defer cancel()
		c := WithContext(ctx)
		for i := 0; i < 10; i++ {
			c.Go(func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			})
		}
		err := c.Wait()
		if !errors.Is(err, tc.err) {
			t.Fatalf("expect: %+v, got: %+v\n", tc.err, err)
		}
	}
}
