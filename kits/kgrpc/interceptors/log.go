package interceptors

import (
	"context"
	"time"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LogOption struct  日志拦截器配置.
type LogOption struct {
	// SlowThreshold 慢请求阈值，超过此时间会记录警告日志.
	SlowThreshold time.Duration
	// SkipMethods 跳过日志记录的方法列表.
	SkipMethods map[string]bool
}

// UnaryLogInterceptor function 一元日志拦截器.
func UnaryLogInterceptor(opts ...func(*LogOption)) grpc.UnaryServerInterceptor {
	option := &LogOption{
		SlowThreshold: 3 * time.Second,
		SkipMethods:   make(map[string]bool),
	}
	for _, opt := range opts {
		opt(option)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if option.SkipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		start := time.Now()
		lg := logger.FromContext(ctx).WithField("method", info.FullMethod)

		resp, err = handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		fields := map[string]interface{}{
			"duration": duration.String(),
			"code":     st.Code().String(),
		}

		if err != nil {
			fields["error"] = err.Error()
			lg.WithFields(fields).Error("gRPC request failed")
		} else if duration > option.SlowThreshold {
			lg.WithFields(fields).Warn("gRPC slow request")
		} else {
			lg.WithFields(fields).Info("gRPC request completed")
		}

		return resp, err
	}
}

// StreamLogInterceptor function  流式日志拦截器.
func StreamLogInterceptor(opts ...func(*LogOption)) grpc.StreamServerInterceptor {
	option := &LogOption{
		SlowThreshold: 3 * time.Second,
		SkipMethods:   make(map[string]bool),
	}
	for _, opt := range opts {
		opt(option)
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if option.SkipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		start := time.Now()
		log := logger.FromContext(ss.Context()).WithField("method", info.FullMethod)

		err := handler(srv, ss)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		fields := map[string]interface{}{
			"duration": duration.String(),
			"code":     st.Code().String(),
		}

		if err != nil {
			fields["error"] = err.Error()
			log.WithFields(fields).Error("gRPC stream failed")
		} else {
			log.WithFields(fields).Info("gRPC stream completed")
		}

		return err
	}
}
