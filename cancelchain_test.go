package cancelchain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGo(t *testing.T) {
	cases := []struct {
		seq    int
		err    error
		expect []int
	}{
		{
			seq:    0,
			err:    errors.New("err"),
			expect: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			seq:    1,
			err:    errors.New("err"),
			expect: []int{1, 0, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			seq:    2,
			err:    errors.New("err"),
			expect: []int{2, 0, 1, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			seq:    3,
			err:    errors.New("err"),
			expect: []int{3, 0, 1, 2, 4, 5, 6, 7, 8, 9},
		},
		{
			seq:    4,
			err:    errors.New("err"),
			expect: []int{4, 0, 1, 2, 3, 5, 6, 7, 8, 9},
		},
		{
			seq:    5,
			expect: []int{5, 0, 1, 2, 3, 4, 6, 7, 8, 9},
		},
		{
			seq:    6,
			err:    errors.New("err"),
			expect: []int{6, 0, 1, 2, 3, 4, 5, 7, 8, 9},
		},
		{
			seq:    7,
			err:    errors.New("err"),
			expect: []int{7, 0, 1, 2, 3, 4, 5, 6, 8, 9},
		},
		{
			seq:    8,
			err:    errors.New("err"),
			expect: []int{8, 0, 1, 2, 3, 4, 5, 6, 7, 9},
		},
		{
			seq:    9,
			err:    errors.New("err"),
			expect: []int{9, 0, 1, 2, 3, 4, 5, 6, 7, 8},
		},
	}

	for _, tc := range cases {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cc := WithContext(ctx)

		got := make([]int, 0, 10)
		for i := 0; i < 10; i++ {
			i := i
			cc.Go(func(ctx context.Context) error {
				if i == tc.seq {
					got = append(got, i)
					return tc.err
				}
				<-ctx.Done()
				got = append(got, i)
				return ctx.Err()
			})
		}

		err := cc.Wait()
		if !assert.Equal(t, tc.expect, got) {
			t.Fatalf("\nexpect %+v\n   got %+v\n", tc.expect, got)
		}
		if !errors.Is(err, tc.err) {
			t.Fatalf("\nexpect %v\n   got %v\n", tc.err, err)
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
