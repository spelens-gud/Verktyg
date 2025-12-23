package kdaemon

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

func Ticker(ctx context.Context, eg *errgroup.Group, seconds int, f func(ctx context.Context), opts ...Option) {
	opt := Options(opts).New()
	if seconds == 0 {
		panic("seconds cannot be 0")
	}
	// 使用 errgroup 进行goroutine计数和获取第一个发生的错误
	eg.Go(func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("%+v", e)
			}
		}()
		ticker := time.NewTicker(time.Second * time.Duration(seconds))
		defer ticker.Stop()
		for doTicker(ctx, ticker, f, opt) {
		}
		return
	})
}

func doTicker(pctx context.Context, ticker *time.Ticker, f func(context.Context), o *opt) (d bool) {
	ctx := itrace.WithContext(pctx, itrace.NewIDFunc())
	if o.IgnorePanic {
		defer func() {
			// 使用 defer recover 保护eg栈 确保能够正常退出
			if e := recover(); e != nil {
				// 处理panic
				if o.OnPanic != nil {
					o.OnPanic(ctx, e)
				}
				d = true
				return
			}
		}()
	}
	select {
	// 定时器信号
	case <-ticker.C:
		for _, h := range o.BeforeHooks {
			h(ctx)
		}
		f(ctx)
		for _, h := range o.AfterHooks {
			h(ctx)
		}
	// 来自上层ctx的中断信号
	case <-ctx.Done():
		for _, h := range o.QuitHooks {
			h(ctx)
		}
		return false
	}
	return true
}
