package gin_middles

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

func TestTrace(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(ExtractTrace())
	e.Any("", func(c *gin.Context) {
		logger.FromContext(c.Request.Context()).Error("")
	})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.Header.Set("X-Forwarded-For", "10.42.2.106")
	e.ServeHTTP(rec, req)
}
