package kserver

import (
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/kits/kserver/httpprof"

	"github.com/spelens-gud/Verktyg/kits/kserver/metrics"

	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kserver/gin_middles"
)

func NewGinEngine() *gin.Engine {
	e := gin.New()

	// 非开发模式
	if !iconfig.GetEnv().IsDevelopment() {
		// 健康检查
		gin_middles.HealthCheck(e)
		// metrics
		metrics.RegisterMetricsExport(e)
		// pprof(使用环境变量控制pprof的启用与否，而非环境标识)
		httpprof.PProf(e)
	}

	// 生产模式
	if iconfig.GetEnv().IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	return e
}
