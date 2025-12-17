package gin_middles

import (
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg.git/implements/promhttp"
	"github.com/spelens-gud/Verktyg.git/interfaces/imetrics"
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
