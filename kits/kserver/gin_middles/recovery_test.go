package gin_middles

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(Recovery())
	e.Any("", func(c *gin.Context) {
		e := errors.New("panic")
		panic(e)
	})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	e.ServeHTTP(rec, req)
}
