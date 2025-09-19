package internal

import (
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
)

var (
	initOnce      sync.Once
	singleton     *provider
	ctxBackground = context.Background()
)

type (
	provider struct {
		sync.Mutex
		instance atomic.Value
		option   *option
	}
)

func getRuntime() *provider {
	initOnce.Do(func() {
		singleton = &provider{option: &option{output: os.Stdout}}
		singleton.update(func(option *option) {})
	})
	return singleton
}

func InitContextLogger(ctx context.Context) (logger ilog.Logger) {
	inst := getRuntime().getInstance()
	if ctx == ctxBackground {
		return inst.GetCtxBackgroundLogger()
	}
	return initContextLogger(ctx, inst.GetLogger())
}

func initContextLogger(ctx context.Context, logger ilog.Logger) ilog.Logger {
	r := getRuntime()

	// 动态全局context补丁
	for _, p := range r.option.contextPatch {
		logger = p.Patch(ctx, logger)
	}

	// 额外注入context补丁
	for _, p := range ExtPatchFromContext(ctx) {
		logger = p.Patch(ctx, logger)
	}

	// level判断
	if level := r.option.level; !level.EnableAll() {
		logger = wrapLogger(logger, levelLog(level).Compare)
	}
	return logger
}

func Flush() {
	getRuntime().getInstance().Flush()
}

func Writer() io.Writer {
	return getRuntime()
}

func (r *provider) getInstance() *instance {
	return r.instance.Load().(*instance)
}

func (r *provider) Write(p []byte) (n int, err error) {
	return r.getInstance().GetWriter().Write(p)
}

func (r *provider) update(f func(option *option)) {
	r.Lock()
	defer r.Unlock()
	f(r.option)
	old, _ := r.instance.Load().(*instance)
	r.instance.Store(&instance{option: r.option})
	if old != nil {
		go old.Flush()
	}
}
