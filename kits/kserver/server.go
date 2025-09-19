package kserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"git.bestfulfill.tech/devops/go-core/kits/klog/bufferedwriter"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
	"git.bestfulfill.tech/devops/go-core/kits/kruntime"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/graceful"
)

const defaultCloseWaitSeconds = 10

func Run(handler http.Handler, serverConfig Config, serverOptions ...func(server *http.Server)) {
	var (
		lg              = logger.FromBackground().WithField("pid", syscall.Getpid())
		config          = &serverConfig
		isGracefulChild = graceful.IsGracefulChild()
		parentPID       = syscall.Getppid()

		listener net.Listener
		err      error
	)

	config.Init()

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           handler,
		ReadTimeout:       time.Duration(config.ReadTimeoutSeconds) * time.Second,
		ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(config.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:       time.Duration(config.IdleTimeoutSeconds) * time.Second,
		ErrorLog:          logger.NewStandardLogger(),
		ConnState:         config.Metrics.SetConnectionsState,
	}

	if config.BaseContext != nil {
		server.BaseContext = func(listener net.Listener) context.Context {
			return config.BaseContext
		}
	}

	if isGracefulChild {
		lg.Infof("listening to socket fd from ppid [ %d ]", parentPID)
		listener, err = net.FileListener(os.NewFile(3, ""))
	} else {
		listener, err = net.Listen("tcp", server.Addr)
	}

	if err != nil {
		lg.Errorf("init listener error: %v", err)
		return
	}

	var (
		wrappedListener = NewAcceptedNoticeListener(listener)
		serveErrChan    = make(chan error)
		stopChan        = kruntime.StopChan()
	)

	if config.EnableAsyncLogger {
		var bypassWriter io.WriteCloser

		logger.RegisterWriterWrapper(func(writer io.Writer) io.Writer {
			return bufferedwriter.NewBypassBufferedWriter(writer, bufferedwriter.NewZapBufferedWriter)
		})

		lg.Infof("update logger to async writer")

		server.RegisterOnShutdown(func() {
			if bypassWriter != nil {
				_ = bypassWriter.Close()
			}
		})
	}

	for _, opt := range serverOptions {
		opt(server)
	}

	go func() {
		lg.Infof("server start listening on [ %v ]", listener.Addr().String())
		serveErrChan <- server.Serve(wrappedListener)
	}()

	if isGracefulChild {
		go func() {
			select {
			case <-wrappedListener.Accepted():
				// 开始服务后 发送信号kill父进程
				time.Sleep(time.Second)
				_ = syscall.Kill(parentPID, syscall.SIGTERM)
			case <-serveErrChan:
			}
		}()
	}

	closeWaitDuration := time.Duration(config.CloseWaitSeconds) * time.Second

	if config.GracefulRestart {
		go func() {
			var lastFork time.Time
			// 收到重启信号 fork创建子进程
			for range kruntime.SignalChan(1, graceful.RestartSignals...) {
				// 最近一次触发fork在 closeWait+1s 时间内 忽略
				if !lastFork.IsZero() && time.Since(lastFork) < closeWaitDuration+time.Second {
					continue
				}

				if childPid, forkErr := graceful.ForkWithListener(listener); forkErr == nil {
					lg.Infof("fork new process success,pid : [ %d ]", childPid)
					lastFork = time.Now()
				} else {
					lg.Errorf("fork new process with listener error: %v", forkErr)
				}
			}
		}()
	}

	select {
	case serveErr := <-serveErrChan:
		lg.Errorf("server serve error: %v", serveErr)
	case <-stopChan:
		server.SetKeepAlivesEnabled(false)

		// 退出前等待处理完所有请求
		shutdownCtx, cancel := context.WithTimeout(context.Background(), closeWaitDuration)
		defer cancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			lg.Errorf("close wait timeout [ %s ],forced to shutdown: %v", closeWaitDuration, shutdownErr)
		}
	}

	lg.Info("server exiting...")
}
