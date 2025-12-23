package gin_middles

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

func init() {
	//gin.SetMode(gin.ReleaseMode)
	//tracer.SetTracer(skytrace.NewLogTracer("test-service", logger.FromBackground()))
}

func TestLogger(t *testing.T) {
	eg := gin.New()
	eg.Use(DefaultChain()...)
	eg.GET("/test", func(c *gin.Context) {
		ctx := c.Request.Context()
		logger.FromContext(ctx).WithTag("MYSQL").WithTag("TEST").Info("test")
		//logger.FromContext(ctx).Error("test")
		//panic("xxx")
	})
	eg.POST("/test", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		logger.FromContext(c.Request.Context()).Infof("%s", b)
		//logger.FromContext(ctx).Error("test")
		//panic("xxx")
	})
	go func() {
		_ = eg.Run(":80")
	}()
	_, err := http.Get("http://localhost/test?ttt=sdgdagda")
	if err != nil {
		t.Fatal(err)
	}
	wg := &errgroup.Group{}
	for i := 0; i < 100; i++ {
		wg.Go(func() error {
			req, _ := http.NewRequest("POST", "http://localhost/test?ttt=sdgdagda", bytes.NewBufferString("test"))
			req.AddCookie(&http.Cookie{
				Name:  "asgdgdsa",
				Value: "gadsghffadsa",
			})
			_, err = http.DefaultClient.Do(req)
			return err
		})
	}
	_ = wg.Wait()
}
