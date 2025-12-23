// Package interceptors 提供 gRPC 服务器拦截器.
// 包含日志、链路追踪、监控、恢复等默认拦截器.
package interceptors

import (
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
	"google.golang.org/grpc"
)

// ChainOption struct    拦截器链配置选项.
type ChainOption struct {
	DisableLog      bool // 禁用日志拦截器
	DisableMetrics  bool // 禁用监控拦截器
	DisableTrace    bool // 禁用链路追踪拦截器
	DisableRecovery bool // 禁用恢复拦截器

	LogOptions      []func(*LogOption)
	TraceOptions    []func(*TraceOption)
	RecoveryOptions []func(*RecoveryOption)
	MetricsOptions  []func(*MetricsOption)
}

// 初始化默认的 ChainOption
func newDefaultChainOption() *ChainOption {
	return &ChainOption{
		DisableMetrics: envflag.IsNotEmpty(iconfig.EnvKeyServerMetricsDisable),
		DisableTrace:   envflag.IsNotEmpty(iconfig.EnvKeyServerTraceDisable),
	}
}

// applyOptions 应用配置选项
func (o *ChainOption) applyOptions(opts ...func(*ChainOption)) {
	for _, opt := range opts {
		opt(o)
	}
}

// appendInterceptors 将未禁用的拦截器添加到链中
func (o *ChainOption) appendInterceptors(chain []grpc.UnaryServerInterceptor) []grpc.UnaryServerInterceptor {
	// 监控指标
	if !o.DisableMetrics {
		chain = append(chain, UnaryMetricsInterceptor(o.MetricsOptions...))
	}

	// 日志
	if !o.DisableLog {
		chain = append(chain, UnaryLogInterceptor(o.LogOptions...))
	}

	// 链路追踪
	if !o.DisableTrace {
		chain = append(chain, UnaryTraceInterceptor(o.TraceOptions...))
	}

	// 恢复（放在最后，捕获所有 panic）
	if !o.DisableRecovery {
		chain = append(chain, UnaryRecoveryInterceptor(o.RecoveryOptions...))
	}

	return chain
}

// appendStreamInterceptors 将未禁用的流式拦截器添加到链中
func (o *ChainOption) appendStreamInterceptors(chain []grpc.StreamServerInterceptor) []grpc.StreamServerInterceptor {
	// 监控指标
	if !o.DisableMetrics {
		chain = append(chain, StreamMetricsInterceptor(o.MetricsOptions...))
	}

	// 日志
	if !o.DisableLog {
		chain = append(chain, StreamLogInterceptor(o.LogOptions...))
	}

	// 链路追踪
	if !o.DisableTrace {
		chain = append(chain, StreamTraceInterceptor(o.TraceOptions...))
	}

	// 恢复
	if !o.DisableRecovery {
		chain = append(chain, StreamRecoveryInterceptor(o.RecoveryOptions...))
	}

	return chain
}

// DefaultUnaryChain function 获取默认的一元拦截器链.
func DefaultUnaryChain(opts ...func(*ChainOption)) []grpc.UnaryServerInterceptor {
	option := newDefaultChainOption()
	option.applyOptions(opts...)

	chain := make([]grpc.UnaryServerInterceptor, 0, 4)
	chain = option.appendInterceptors(chain)

	return chain
}

// DefaultStreamChain function 获取默认的流式拦截器链.
func DefaultStreamChain(opts ...func(*ChainOption)) []grpc.StreamServerInterceptor {
	option := newDefaultChainOption()
	option.applyOptions(opts...)

	chain := make([]grpc.StreamServerInterceptor, 0, 4)
	chain = option.appendStreamInterceptors(chain)

	return chain
}
