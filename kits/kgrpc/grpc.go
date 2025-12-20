package kgrpc

import (
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kgrpc/interceptors"
	"google.golang.org/grpc"
)

// NewGrpcServerWithInterceptors function 创建带自定义拦截器的 gRPC 服务器.
func NewGrpcServerWithInterceptors(
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor,
	opts ...grpc.ServerOption,
) *grpc.Server {
	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(defaultMaxMsgSize),
		grpc.MaxSendMsgSize(defaultMaxMsgSize),
	}

	if len(unaryInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}

	if len(streamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	serverOpts = append(serverOpts, opts...)

	return grpc.NewServer(serverOpts...)
}

// NewGrpcServer function 创建带默认配置的 gRPC 服务器.
func NewGrpcServer(opts ...grpc.ServerOption) *grpc.Server {
	// 默认选项
	defaultOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(defaultMaxMsgSize),
		grpc.MaxSendMsgSize(defaultMaxMsgSize),
	}

	// 非开发模式添加默认拦截器
	if !iconfig.GetEnv().IsDevelopment() {
		defaultOpts = append(defaultOpts,
			grpc.ChainUnaryInterceptor(interceptors.DefaultUnaryChain()...),
			grpc.ChainStreamInterceptor(interceptors.DefaultStreamChain()...),
		)
	}

	// 合并用户选项（用户选项优先）
	allOpts := append(defaultOpts, opts...)

	return grpc.NewServer(allOpts...)
}
