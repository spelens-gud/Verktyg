package kdaemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

func StartDaemon(ctx context.Context, startFunc func(context.Context, *errgroup.Group)) {
	// 中断信号
	ctx, cancelFunc := context.WithCancel(ctx)

	// 创建errgruop
	eg, ctx := errgroup.WithContext(ctx)

	logger.FromContext(ctx).Info("init worker")

	// 将包含中断信号的ctx和errgroup提供给注册方法
	startFunc(ctx, eg)

	// 系统终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	logger.FromContext(ctx).Info("start worker")

	go func() {
		exitCode := 0
		// 挂起任务
		err := eg.Wait()
		// 异常退出
		if err != nil {
			exitCode = 1
			logger.FromContext(ctx).Errorf("daemon error: %+v", err)
		}
		// 所有任务已经退出
		os.Exit(exitCode)
	}()

	// 系统信号退出
	<-quit
	// 发送中断信号
	cancelFunc()
	// 等待10秒未退出的任务
	time.Sleep(time.Second * 10)
}
