package gin_middles

import (
	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/implements/promhttp"
	"git.bestfulfill.tech/devops/go-core/interfaces/imetrics"
)

type MetricsOption struct {
	Metrics imetrics.ServerMetrics
}

func Metrics(opts ...func(option *MetricsOption)) gin.HandlerFunc {
	opt := &MetricsOption{
		Metrics: promhttp.DefaultServerMetrics,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt.Metrics.GinInterceptor()
}
