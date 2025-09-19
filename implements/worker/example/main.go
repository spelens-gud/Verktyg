package main

import (
	"context"
	"errors"

	"github.com/robfig/cron/v3"

	"git.bestfulfill.tech/devops/go-core/implements/worker"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

func main() {
	wk := worker.NewWorkerManager(worker.WithCronOption(cron.WithSeconds()))
	wk.RegisterCron("* * * * * *", func(ctx context.Context) (err error) {
		logger.FromContext(ctx).Info("test")
		return nil
	})
	wk.RegisterCron("* * * * * *", func(ctx context.Context) (err error) {
		logger.FromContext(ctx).Info("test")
		return errors.New("test")
	})
	wk.RegisterTasks(func(ctx context.Context) {
		<-ctx.Done()
		logger.FromContext(ctx).Info("task exit")
	})
	wk.RegisterTasks(func(ctx context.Context) {
		<-ctx.Done()
		logger.FromContext(ctx).Info("task2 exit")
	})
	wk.RegisterTasks(func(ctx context.Context) {
		<-ctx.Done()
		logger.FromContext(ctx).Info("task3 exit")
	})
	wk.Start()
}
