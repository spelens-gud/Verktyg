package kserver

import (
	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/kits/kserver/httpprof"

	"git.bestfulfill.tech/devops/go-core/kits/kserver/metrics"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/gin_middles"
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
