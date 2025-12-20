package kgrpc

import (
	"context"
	"net"
	"syscall"
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/ilog"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
	"github.com/spelens-gud/Verktyg/kits/kruntime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Run function 启动 gRPC 服务器.
func Run(server *grpc.Server, config Config) {
	log := logger.FromBackground().WithField("pid", syscall.Getpid())
	stopChan := kruntime.StopChan()
	runInternal(server, config, stopChan, log)
}

// RunWithContext function 带上下文启动 gRPC 服务器.
func RunWithContext(ctx context.Context, server *grpc.Server, config Config) {
	log := logger.FromBackground().WithField("pid", syscall.Getpid())
	ctxDone := ctx.Done()
	runInternal(server, config, ctxDone, log)
}

// runInternal 是内部函数，封装了 gRPC 服务器运行的核心逻辑
func runInternal(server *grpc.Server, config Config, stopCh <-chan struct{}, log ilog.Logger) {
	config.Init()

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		log.Errorf("gRPC server failed to listen: %v", err)
		return
	}

	// 启用反射服务
	if config.EnableReflection {
		reflection.Register(server)
		log.Info("gRPC reflection service enabled")
	}

	serveErrChan := make(chan error)

	go func() {
		log.Infof("gRPC server start listening on [ %v ]", listener.Addr().String())
		serveErrChan <- server.Serve(listener)
	}()

	select {
	case serveErr := <-serveErrChan:
		log.Errorf("gRPC server serve error: %v", serveErr)
	case <-stopCh:
		log.Info("gRPC server shutting down...")

		// 优雅关闭
		done := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Info("gRPC server graceful shutdown completed")
		case <-time.After(config.GetCloseWaitDuration()):
			log.Warn("gRPC server graceful shutdown timeout, forcing stop")
			server.Stop()
		}
	}

	log.Info("gRPC server exiting...")
}
