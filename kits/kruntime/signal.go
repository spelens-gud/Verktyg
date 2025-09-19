package kruntime

import (
	"os"
	"os/signal"
	"sync"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

var (
	stopCh   <-chan struct{}
	initOnce sync.Once
)

func SignalChan(size int, sig ...os.Signal) <-chan os.Signal {
	c := make(chan os.Signal, size)
	signal.Notify(c, sig...)
	return c
}

func initStopChan() {
	initOnce.Do(func() {
		var (
			stop = make(chan struct{})
			c    = SignalChan(2, shutdownSignals...)
			lg   = logger.FromBackground().WithField("pid", os.Getpid())
		)
		lg.Infof("stop signal listening: %v", shutdownSignals)
		go func() {
			sig := <-c
			lg.Infof("stop signal received: [ %s ]", sig.String())
			close(stop)
			sig = <-c
			lg.Infof("stop signal received again,force exit: [ %s ]", sig.String())
			os.Exit(1) // second signal. Exit directly.
		}()
		stopCh = stop
	})
}

func StopChan() <-chan struct{} {
	initStopChan()
	return stopCh
}
