package govern_server

import (
	"os"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/kenv/envflag"
	"git.bestfulfill.tech/devops/go-core/kits/kgo"
	"git.bestfulfill.tech/devops/go-core/kits/kserver"
)

func RunGovernServer() {
	// 环境变量判断是否需要开启监控采集server
	if envflag.IsEmpty(iconfig.EnvKeyGovernServerEnable) {
		return
	}
	// 环境变量判断开启端口
	eg := kserver.NewGinEngine()
	port := kgo.InitStrings(os.Getenv(iconfig.EnvKeyGovernServerPort), ":7799")
	go func() {
		if err := eg.Run(port); err != nil {
			panic(err)
		}
	}()
}
