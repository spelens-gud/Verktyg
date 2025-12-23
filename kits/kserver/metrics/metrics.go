package metrics

import (
	"os"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/implements/promhttp"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/imetrics"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

var (
	l sync.Mutex

	registeredMap = map[gin.IRouter]bool{}
)

type Option struct {
	MetricsExportPath string
	BasicAuthSecret   string
	Metrics           imetrics.ServerMetrics
}

func RegisterMetricsExport(router gin.IRouter, opts ...func(option *Option)) {
	l.Lock()
	defer l.Unlock()
	if registeredMap[router] {
		return
	}
	registeredMap[router] = true

	opt := &Option{
		Metrics:           promhttp.DefaultServerMetrics,
		MetricsExportPath: os.Getenv(iconfig.EnvKeyRuntimeMetricsExportPath),
		BasicAuthSecret:   os.Getenv(iconfig.EnvKeyRuntimeMetricsExportSecretKey),
	}

	for _, o := range opts {
		o(opt)
	}

	if len(opt.MetricsExportPath) == 0 {
		opt.MetricsExportPath = "/metrics"
	}

	router = router.Group(opt.MetricsExportPath)

	logger.FromBackground().Warnf("metrics api enable on route: [ %s ]", opt.MetricsExportPath)

	if len(opt.BasicAuthSecret) > 0 {
		// 添加basicAuth
		router.Use(gin.BasicAuth(gin.Accounts{"metrics": opt.BasicAuthSecret}))
	}

	// 暴露采集路径
	handler := gin.WrapF(opt.Metrics.Handler())
	router.Any("", handler)
	router.Any("/", handler)
}
