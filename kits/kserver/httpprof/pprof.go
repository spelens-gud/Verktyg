package httpprof

import (
	"os"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/kenv/envflag"
	"git.bestfulfill.tech/devops/go-core/kits/kgo"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

func PProf(group gin.IRouter) {
	if envflag.IsEmpty(iconfig.EnvKeyRuntimePprofEnable) {
		return
	}

	var (
		// 获取前缀
		groupPrefix = kgo.InitStrings(os.Getenv(iconfig.EnvKeyRuntimePprofPrefix), pprof.DefaultPrefix)
		// 路由组
		pprofGroup = group.Group(groupPrefix)
		// basicAuth
		secret = os.Getenv(iconfig.EnvKeyRuntimePprofSecretKey)
	)

	// 日志提醒
	logger.FromBackground().Warnf("pprof debug enable on route: [ %s ]", groupPrefix)

	if len(secret) > 0 {
		pprofGroup.Use(gin.BasicAuth(gin.Accounts{"pprof": secret}))
	}

	pprof.RouteRegister(pprofGroup, "")
}
