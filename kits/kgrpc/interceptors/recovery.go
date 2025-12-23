package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryOption struct 恢复拦截器配置.
type RecoveryOption struct {
	// RecoveryHandler 自定义恢复处理函数.
	RecoveryHandler func(ctx context.Context, p any) error
}

// UnaryRecoveryInterceptor function 一元恢复拦截器.
func UnaryRecoveryInterceptor(opts ...func(*RecoveryOption)) grpc.UnaryServerInterceptor {
	option := &RecoveryOption{}
	for _, opt := range opts {
		opt(option)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				lg := logger.FromContext(ctx)
				lg.Errorf("gRPC panic recovered: %v\n%s", p, debug.Stack())

				if option.RecoveryHandler != nil {
					err = option.RecoveryHandler(ctx, p)
				} else {
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor function 流式恢复拦截器.
func StreamRecoveryInterceptor(opts ...func(*RecoveryOption)) grpc.StreamServerInterceptor {
	option := &RecoveryOption{}
	for _, opt := range opts {
		opt(option)
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if p := recover(); p != nil {
				lg := logger.FromContext(ss.Context())
				lg.Errorf("gRPC stream panic recovered: %v\n%s", p, debug.Stack())

				if option.RecoveryHandler != nil {
					err = option.RecoveryHandler(ss.Context(), p)
				} else {
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}
		}()

		return handler(srv, ss)
	}
}
