package gin_middles

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg.git/kits/kcontext"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

func TestMetadata(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(ExtractMetadata())
	e.Any("", func(c *gin.Context) {
		logger.FromContext(c.Request.Context()).Infof("%v", kcontext.GetMetadata(c.Request.Context()))
	})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.Header.Add("test", "test2")
	e.ServeHTTP(rec, req)
}
