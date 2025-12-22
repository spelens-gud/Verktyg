package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsOption struct 监控指标拦截器配置.
type MetricsOption struct {
	// Collector 指标收集器.
	Collector MetricsCollector
}

// MetricsCollector interface 指标收集器接口.
type MetricsCollector interface {
	// RecordRequest 记录请求指标.
	RecordRequest(method string, code string, duration time.Duration)
}

// defaultMetricsCollector struct 默认指标收集器（空实现）.
type defaultMetricsCollector struct{}

func (d *defaultMetricsCollector) RecordRequest(method string, code string, duration time.Duration) {
	// 默认空实现，可以通过选项注入实际的 Prometheus 收集器
}

// UnaryMetricsInterceptor function 一元监控指标拦截器.
func UnaryMetricsInterceptor(opts ...func(*MetricsOption)) grpc.UnaryServerInterceptor {
	option := &MetricsOption{
		Collector: &defaultMetricsCollector{},
	}
	for _, opt := range opts {
		opt(option)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		option.Collector.RecordRequest(info.FullMethod, st.Code().String(), duration)

		return resp, err
	}
}

// StreamMetricsInterceptor function 流式监控指标拦截器.
func StreamMetricsInterceptor(opts ...func(*MetricsOption)) grpc.StreamServerInterceptor {
	option := &MetricsOption{
		Collector: &defaultMetricsCollector{},
	}
	for _, opt := range opts {
		opt(option)
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		option.Collector.RecordRequest(info.FullMethod, st.Code().String(), duration)

		return err
	}
}
