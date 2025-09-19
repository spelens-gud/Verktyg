package worker

import (
	"context"

	"github.com/robfig/cron/v3"

	"git.bestfulfill.tech/devops/go-core/interfaces/iworker"
)

var _ iworker.Worker = &worker{}

type worker struct {
	iworker.WorkerManager
}

// Deprecated: 使用 NewWorkerManager
func NewWorker(options ...cron.Option) iworker.Worker {
	return &worker{
		WorkerManager: NewWorkerManager(func(option *Option) {
			option.CronOptions = options
		}),
	}
}

func (w worker) OnCancel(cancelFunc ...func()) {
	w.RegisterTasks(func(ctx context.Context) {
		<-ctx.Done()
		for _, cancel := range cancelFunc {
			cancel()
		}
	})
}

func (w worker) Run() {
	// 异步启动治理端口
	w.Start()
}

func (w worker) StartCron(cronSpec string, cmdFunc func(ctx context.Context) (err error)) {
	w.RegisterCron(cronSpec, cmdFunc)
}
