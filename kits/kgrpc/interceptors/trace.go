package interceptors

import (
	"context"

	"github.com/spelens-gud/Verktyg/kits/ktrace/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TraceOption struct 链路追踪拦截器配置.
type TraceOption struct{}

// UnaryTraceInterceptor function 一元链路追踪拦截器.
func UnaryTraceInterceptor(opts ...func(*TraceOption)) grpc.UnaryServerInterceptor {
	option := &TraceOption{}
	for _, opt := range opts {
		opt(option)
	}
	_ = option // 预留扩展

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 从 metadata 中提取并创建 span
		md := make(map[string][]string)
		if inMd, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range inMd {
				md[k] = v
			}
		}

		span, ctx := tracer.ExtractMetadata(ctx, info.FullMethod, md)
		defer span.Finish()

		return handler(ctx, req)
	}
}

// StreamTraceInterceptor function 流式链路追踪拦截器.
func StreamTraceInterceptor(opts ...func(*TraceOption)) grpc.StreamServerInterceptor {
	option := &TraceOption{}
	for _, opt := range opts {
		opt(option)
	}
	_ = option // 预留扩展

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		md := make(map[string][]string)
		if inMd, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range inMd {
				md[k] = v
			}
		}

		span, ctx := tracer.ExtractMetadata(ctx, info.FullMethod, md)
		defer span.Finish()

		// 包装 ServerStream 以使用新的上下文
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// wrappedServerStream struct 包装的 ServerStream，用于注入上下文.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context method 返回包装的上下文.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
