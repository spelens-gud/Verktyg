package govern_server

import (
	"os"

	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/kgo"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/gin_middles"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/httpprof"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/metrics"
)

func RunGovernServer(ports ...string) {
	// 环境变量判断开启端口
	eg := gin.New()

	// 健康检查
	gin_middles.HealthCheck(eg)
	// metrics
	metrics.RegisterMetricsExport(eg)

	// pprof(使用环境变量控制pprof的启用与否，而非环境标识)
	httpprof.PProf(eg)

	// 生产模式
	if iconfig.GetEnv().IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 端口
	ports = append(ports, os.Getenv(iconfig.EnvKeyGovernServerPort), ":7799")

	port := kgo.InitStrings(ports...)
	go func() {
		if err := eg.Run(port); err != nil {
			panic(err)
		}
	}()
}
