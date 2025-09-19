package iworker

import "context"

type WorkerManager interface {
	Start()

	RegisterTasks(f ...func(ctx context.Context))

	MustRegisterTask(f func(ctx context.Context), err error)

	RegisterCron(cronSpec string, cmdFunc func(ctx context.Context) (err error))
}

// deprecated: 使用 WorkerManager
type Worker interface {
	OnCancel(cancelFunc ...func())

	Run()

	StartCron(cronSpec string, cmdFunc func(ctx context.Context) (err error))
}
