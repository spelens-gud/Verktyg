// Package interceptors 提供 gRPC 服务器拦截器.
// 包含日志、链路追踪、监控、恢复等默认拦截器.
package interceptors

import (
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
	"google.golang.org/grpc"
)

// ChainOption struct 拦截器链配置选项.
type ChainOption struct {
	DisableLog      bool // 是否禁用日志
	DisableMetrics  bool // 是否禁用监控
	DisableTrace    bool // 是否禁用链路追踪
	DisableRecovery bool // 是否禁用恢复

	LogOptions      []func(*LogOption)
	TraceOptions    []func(*TraceOption)
	RecoveryOptions []func(*RecoveryOption)
	MetricsOptions  []func(*MetricsOption)
}

// 初始化 ChainOption 的通用函数
func initChainOption() *ChainOption {
	return &ChainOption{
		DisableMetrics: envflag.IsNotEmpty(iconfig.EnvKeyServerMetricsDisable),
		DisableTrace:   envflag.IsNotEmpty(iconfig.EnvKeyServerTraceDisable),
	}
}

// applyOptions 应用配置选项
func applyOptions(option *ChainOption, opts ...func(*ChainOption)) {
	for _, opt := range opts {
		opt(option)
	}
}

// DefaultUnaryChain function 获取默认的一元拦截器链.
func DefaultUnaryChain(opts ...func(*ChainOption)) []grpc.UnaryServerInterceptor {
	option := initChainOption()
	applyOptions(option, opts...)

	chain := make([]grpc.UnaryServerInterceptor, 0, 4)

	// 监控指标
	if !option.DisableMetrics {
		chain = append(chain, UnaryMetricsInterceptor(option.MetricsOptions...))
	}

	// 日志
	if !option.DisableLog {
		chain = append(chain, UnaryLogInterceptor(option.LogOptions...))
	}

	// 链路追踪
	if !option.DisableTrace {
		chain = append(chain, UnaryTraceInterceptor(option.TraceOptions...))
	}

	// 恢复（放在最后，捕获所有 panic）
	if !option.DisableRecovery {
		chain = append(chain, UnaryRecoveryInterceptor(option.RecoveryOptions...))
	}

	return chain
}

// DefaultStreamChain function 获取默认的流式拦截器链.
func DefaultStreamChain(opts ...func(*ChainOption)) []grpc.StreamServerInterceptor {
	option := initChainOption()
	applyOptions(option, opts...)

	chain := make([]grpc.StreamServerInterceptor, 0, 4)

	// 监控指标
	if !option.DisableMetrics {
		chain = append(chain, StreamMetricsInterceptor(option.MetricsOptions...))
	}

	// 日志
	if !option.DisableLog {
		chain = append(chain, StreamLogInterceptor(option.LogOptions...))
	}

	// 链路追踪
	if !option.DisableTrace {
		chain = append(chain, StreamTraceInterceptor(option.TraceOptions...))
	}

	// 恢复
	if !option.DisableRecovery {
		chain = append(chain, StreamRecoveryInterceptor(option.RecoveryOptions...))
	}

	return chain
}
