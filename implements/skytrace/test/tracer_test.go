package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/implements/redisx"
	"git.bestfulfill.tech/devops/go-core/implements/skytrace"
	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/interfaces/iredis"
	"git.bestfulfill.tech/devops/go-core/kits/kcontext"
	"git.bestfulfill.tech/devops/go-core/kits/kserver"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/gin_middles"
	"git.bestfulfill.tech/devops/go-core/kits/ktrace/tracer"
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
		rd.GetRedis().Get(ctx, "key")
		defer sp.Finish()
	})
	_ = e.Run(":80")
}
