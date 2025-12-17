package kserver

import (
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/spelens-gud/Verktyg.git/implements/httpreq"
	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg.git/interfaces/imetrics"
	"github.com/spelens-gud/Verktyg.git/kits/kserver/gin_middles"
	"github.com/spelens-gud/Verktyg.git/kits/kserver/metrics"
)

func init() {
	//tc, _, _ := skytrace.NewGrpcTracer("demo-service", "127.0.0.1:11800", reporter.WithLogger(log.New(&mockWriter{}, "", 0)))
	//tracer.SetTracer(tc)
}

func TestServer(t *testing.T) {
	e := gin.New()
	metrics.RegisterMetricsExport(e)
	e.Use(gin_middles.DefaultChain()...)
	//e.Use(gin.Recovery())
	cli := httpreq.NewFormClient("http://localhost")
	e.GET("/", func(c *gin.Context) {
		_, err := cli.Get(c.Request.Context(), "/2", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
		c.JSON(http.StatusOK, nil)
	})
	e.GET("/2", func(c *gin.Context) {
		_, err := cli.Get(c.Request.Context(), "/3", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
		c.JSON(http.StatusOK, nil)
	})
	e.GET("/3", func(c *gin.Context) {
		_, err := cli.Get(c.Request.Context(), "/4", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
		c.JSON(http.StatusOK, nil)
	})
	e.GET("/4", func(c *gin.Context) {
		if rand.Intn(10) > 5 {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
		c.JSON(http.StatusOK, nil)
	})
	err := e.Run(":80")
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerCli(t *testing.T) {
	eg := &errgroup.Group{}
	cli := ihttp.NewDefaultHttpClient()
	var v int64
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Duration(rand.Intn(500)) * time.Microsecond)
		eg.Go(func() error {
			resp, err := http.Get("http://localhost")
			if err != nil {
				return err
			}
			v += 1
			_ = resp.Body.Close()
			return nil
		})
	}
	_ = eg.Wait()
	//if err != nil {
	//	t.Fatal(err)
	//}
	print(v)
	resp, err := cli.Get("http://localhost" + imetrics.DefaultMetricsPath)
	if err != nil {
		t.Fatal(err)
	}
	// nolint
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	t.Logf("%s", b)
	time.Sleep(time.Second)
}
