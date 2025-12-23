package gin_middles

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
)

// deprecated: 应改用  gin_middles.DefaultChain()... 代替
// var Default = DefaultChain()

var (
	defaultPanicStatus  = http.StatusInternalServerError
	defaultClientIPFunc = func(c *gin.Context) string {
		return c.ClientIP()
	}
)

type ChainOption struct {
	ClientIpFunc func(c *gin.Context) string
	PanicStatus  int

	DisableLog      bool
	DisableMetrics  bool
	DisableTrace    bool
	DisableMetadata bool
	DisableRecovery bool
	DisableSentinel bool

	LogOptions      []func(*LogOption)
	TraceOptions    []func(*TraceOption)
	RecoveryOption  []func(*RecoveryOption)
	SentinelOptions []func(*SentinelOption)
	MetricsOptions  []func(*MetricsOption)
}

func WithClientIpFunc(f func(c *gin.Context) string) func(patch *ChainOption) {
	return func(patch *ChainOption) {
		patch.ClientIpFunc = f
	}
}

// 默认中间件
func DefaultChain(opts ...func(option *ChainOption)) gin.HandlersChain {
	option := &ChainOption{
		DisableSentinel: iconfig.GetEnv().IsDevelopment() || envflag.IsNotEmpty(iconfig.EnvKeyServerSentinelDisable),
		DisableMetrics:  envflag.IsNotEmpty(iconfig.EnvKeyServerMetricsDisable),
		DisableTrace:    envflag.IsNotEmpty(iconfig.EnvKeyServerTraceDisable),
		ClientIpFunc:    defaultClientIPFunc,
		PanicStatus:     defaultPanicStatus,
	}

	for _, p := range opts {
		p(option)
	}

	defaultPanicStatus = option.PanicStatus
	defaultClientIPFunc = option.ClientIpFunc

	chain := make(gin.HandlersChain, 0, 5)

	// 服务器监控
	if !option.DisableMetrics {
		chain = append(chain, Metrics(option.MetricsOptions...))
	}

	// 自动限流组件
	if !option.DisableSentinel {
		chain = append(chain, Sentinel(option.SentinelOptions...))
	}

	// 服务器请求日志
	if !option.DisableLog {
		chain = append(chain, GinLogger(option.LogOptions...))
	}

	// 链路追踪
	if !option.DisableTrace {
		chain = append(chain, ExtractTrace(option.TraceOptions...))
	}

	// 链路追踪后recovery 捕获服务层panic 会输出栈到链路追踪
	if !option.DisableRecovery {
		chain = append(chain, Recovery(option.RecoveryOption...))
	}

	// 提取头部信息
	if !option.DisableMetadata {
		chain = append(chain, ExtractMetadata())
	}

	return chain
}
