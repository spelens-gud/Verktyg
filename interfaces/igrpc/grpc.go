package igrpc

import "google.golang.org/grpc"

// ServiceRegistrar interface gRPC 服务注册器接口.
type ServiceRegistrar interface {
	// RegisterService 注册 gRPC 服务到服务器.
	RegisterService(server *grpc.Server)
}

// ServiceRegistrarFunc type 服务注册函数类型.
type ServiceRegistrarFunc func(server *grpc.Server)

// RegisterService method 实现 ServiceRegistrar 接口.
func (f ServiceRegistrarFunc) RegisterService(server *grpc.Server) {
	f(server)
}
