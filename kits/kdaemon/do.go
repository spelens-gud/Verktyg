package kdaemon

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

func Always(ctx context.Context, eg *errgroup.Group, f func(ctx context.Context), opts ...Option) {
	opt := Options(opts).New()
	// 使用 errgroup 进行goroutine计数和获取第一个发生的错误
	eg.Go(func() (err error) {
		// 使用 defer recover 保护eg栈 确保能够正常退出
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("%+v", e)
			}
		}()
		// 循环执行
		for doAlways(ctx, f, opt) {
		}
		return
	})
}

func doAlways(pctx context.Context, f func(context.Context), o *opt) (d bool) {
	// 每个循环周期都应该有唯一的请求ID
	ctx := itrace.WithContext(pctx, itrace.NewIDFunc())

	// 无视panic 即遇到panic不退出循环
	if o.IgnorePanic {
		// 使用defer recover保护当前循环运行栈
		defer func() {
			if e := recover(); e != nil {
				// 处理panic
				if o.OnPanic != nil {
					o.OnPanic(ctx, e)
				}
				// 依然返回 true 确保循环继续进行
				d = true
				return
			}
		}()
	}
	select {
	// 来自上层ctx的中断信号
	case <-ctx.Done():
		// 退出前钩子
		for _, h := range o.QuitHooks {
			h(ctx)
		}
		return false
	// 未中断 直接执行流程
	default:
		// 处理前钩子
		for _, h := range o.BeforeHooks {
			h(ctx)
		}
		f(ctx)
		// 处理后钩子
		for _, h := range o.AfterHooks {
			h(ctx)
		}
	}
	return true
}
