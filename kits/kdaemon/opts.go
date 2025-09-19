package kdaemon

import "context"

type (
	opt struct {
		BeforeHooks []func(ctx context.Context)
		AfterHooks  []func(ctx context.Context)
		QuitHooks   []func(ctx context.Context)
		OnPanic     func(ctx context.Context, e interface{})
		IgnorePanic bool
	}

	Option func(*opt)

	Options []Option
)

// 执行任务单元前钩子
func WithBeforeHook(h func(ctx context.Context)) Option {
	return func(o *opt) {
		o.BeforeHooks = append(o.BeforeHooks, h)
	}
}

// 执行任务单元后钩子
func WithAfterHook(h func(ctx context.Context)) Option {
	return func(o *opt) {
		o.AfterHooks = append(o.AfterHooks, h)
	}
}

// 退出任务循环前钩子
func WithQuitHook(h func(ctx context.Context)) Option {
	return func(o *opt) {
		o.QuitHooks = append(o.QuitHooks, h)
	}
}

// 退出任务循环前钩子
func WithPanicHandler(h func(ctx context.Context, e interface{})) Option {
	return func(o *opt) {
		o.OnPanic = h
	}
}

// 遇到panic依然继续任务
func IgnorePanic() Option {
	return func(o *opt) {
		o.IgnorePanic = true
	}
}

func (opts Options) New() *opt {
	o := new(opt)
	for _, op := range opts {
		op(o)
	}
	return o
}
