package gin_middles

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDefault(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(DefaultChain()...)
	e.Any("", func(c *gin.Context) {
		e := errors.New("panic")
		panic(e)
	})
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}
