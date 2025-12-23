package kgrpc

import (
	"context"
	"net"
	"syscall"
	"time"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
	"github.com/spelens-gud/Verktyg/kits/kruntime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Run function 启动 gRPC 服务器.
func Run(server *grpc.Server, config Config) {
	var (
		lg  = logger.FromBackground().WithField("pid", syscall.Getpid())
		err error
	)

	config.Init()

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		lg.Errorf("gRPC server failed to listen: %v", err)
		return
	}

	// 启用反射服务（用于 grpcurl 等工具调试）
	if config.EnableReflection {
		reflection.Register(server)
		lg.Info("gRPC reflection service enabled")
	}

	var (
		serveErrChan = make(chan error)
		stopChan     = kruntime.StopChan()
	)

	go func() {
		lg.Infof("gRPC server start listening on [ %v ]", listener.Addr().String())
		serveErrChan <- server.Serve(listener)
	}()

	select {
	case serveErr := <-serveErrChan:
		lg.Errorf("gRPC server serve error: %v", serveErr)
	case <-stopChan:
		lg.Info("gRPC server shutting down...")

		done := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			lg.Info("gRPC server graceful shutdown completed")
		case <-time.After(config.GetCloseWaitDuration()):
			lg.Warn("gRPC server graceful shutdown timeout, forcing stop")
			server.Stop()
		}
	}

	lg.Info("gRPC server exiting...")
}

// RunWithContext function 带上下文启动 gRPC 服务器.
func RunWithContext(ctx context.Context, server *grpc.Server, config Config) {
	var (
		lg  = logger.FromBackground().WithField("pid", syscall.Getpid())
		err error
	)

	config.Init()

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		lg.Errorf("gRPC server failed to listen: %v", err)
		return
	}

	// 启用反射服务
	if config.EnableReflection {
		reflection.Register(server)
		lg.Info("gRPC reflection service enabled")
	}

	serveErrChan := make(chan error)

	go func() {
		lg.Infof("gRPC server start listening on [ %v ]", listener.Addr().String())
		serveErrChan <- server.Serve(listener)
	}()

	select {
	case serveErr := <-serveErrChan:
		lg.Errorf("gRPC server serve error: %v", serveErr)
	case <-ctx.Done():
		lg.Info("gRPC server shutting down...")

		done := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			lg.Info("gRPC server graceful shutdown completed")
		case <-time.After(config.GetCloseWaitDuration()):
			lg.Warn("gRPC server graceful shutdown timeout, forcing stop")
			server.Stop()
		}
	}

	lg.Info("gRPC server exiting...")
}
