package gin_middles

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func TestSentinel(t *testing.T) {
	r := httptest.NewRecorder()
	g := gin.New()
	g.Use(Sentinel())
	g.GET("", func(c *gin.Context) {

	})
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	var e errgroup.Group
	for i := 0; i < 100; i++ {
		e.Go(func() error {
			g.ServeHTTP(r, req)
			fmt.Println(r.Result().StatusCode)
			return nil
		})
	}
	_ = e.Wait()
}
