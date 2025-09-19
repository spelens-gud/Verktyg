package gin_middles

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
)

var (
	defaultHealthCheckMethod = map[string]bool{http.MethodHead: true}
	defaultHealthCheckRoute  = "/"

	initHealCheckFromEnv sync.Once
)

func HealthCheck(router gin.IRouter) {
	methods, path := getHealthCheckPath()
	for method := range methods {
		router.Handle(method, path, func(c *gin.Context) {
			c.AbortWithStatus(http.StatusOK)
		})
	}
}

func getHealthCheckPath() (method map[string]bool, path string) {
	initHealCheckFromEnv.Do(func() {
		if sp := strings.Split(os.Getenv(iconfig.EnvKeyServerHealthCheckRoute), ":"); len(sp) == 2 {
			for _, m := range strings.Split(sp[0], ",") {
				defaultHealthCheckMethod[m] = true
			}
			defaultHealthCheckRoute = sp[1]
		}
	})
	return defaultHealthCheckMethod, defaultHealthCheckRoute
}
