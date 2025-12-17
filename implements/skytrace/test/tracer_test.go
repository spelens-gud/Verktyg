package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/implements/redisx"
	"github.com/spelens-gud/Verktyg/implements/skytrace"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/iredis"
	"github.com/spelens-gud/Verktyg/kits/kcontext"
	"github.com/spelens-gud/Verktyg/kits/kserver"
	"github.com/spelens-gud/Verktyg/kits/kserver/gin_middles"
	"github.com/spelens-gud/Verktyg/kits/ktrace/tracer"
)

func TestExtractRequestID(t *testing.T) {
	tr, _ := skytrace.NewGrpcTracer("testxxx", "localhost:11800", func(opt *skytrace.TraceOption) {
		opt.SampleRate = 1
	})
	// nolint
	defer tr.Close()

	for i := 0; i < 100; i++ {
		ctx := kcontext.WithNewRequestID(context.Background())
		header := map[string][]string{}
		sp := tr.InjectMetadata(ctx, "test-in", "no", header)
		defer sp.Finish()
		sp2, ctx := tr.ExtractFromMetadata(ctx, "test", header)
		t.Log(go2sky.TraceID(ctx))
		sp2.Finish()
	}

	defer time.Sleep(time.Second)
}

func BenchmarkSpan(b *testing.B) {
	tr, err := skytrace.NewGrpcTracer("testxxx", "localhost:11800")
	if err != nil {
		b.Fatal(err)
	}
	_ = os.Setenv(iconfig.EnvKeyRuntimeMetricsEnable, "1")
	tracer.SetTracer(tr)
	// nolint
	defer tr.Close()
	e := kserver.NewGinEngine()
	e.Use(gin_middles.DefaultChain()...)
	rd, err := redisx.NewRedis(&iredis.RedisConfig{})
	if err != nil {
		b.Fatal(err)
	}
	// nolint
	defer rd.Close()
	pprof.RouteRegister(e.Group(""))

	e.Any("", func(c *gin.Context) {
		sp, ctx := tr.StartSpan(c.Request.Context(), "in")
		rd.GetRedis(context.Background()).Get(ctx, "key")
		defer sp.Finish()
	})
	_ = e.Run(":80")
}
