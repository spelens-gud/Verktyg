package kserver

import (
	"net/http"
	"testing"
	"time"

	sentinelgin "github.com/alibaba/sentinel-golang/adapter/gin"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"

	"git.bestfulfill.tech/devops/go-core/implements/otrace"
	"git.bestfulfill.tech/devops/go-core/interfaces/itrace"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/gin_middles"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/metrics"
	"git.bestfulfill.tech/devops/go-core/kits/ktrace/tracer"
)

var (
	tc itrace.Tracer

	useTrace = ""
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	switch useTrace {
	case "jaeger":
		t, _, _ := jaegerConfig.Configuration{
			ServiceName: "ffff",
			Reporter: &jaegerConfig.ReporterConfig{
				LocalAgentHostPort: "localhost:6831",
			},
		}.NewTracer(
			jaegerConfig.Sampler(jaeger.NewConstSampler(true)),
		)
		opentracing.SetGlobalTracer(t)
		tc = &otrace.Tracer{}
	}
}

type mockWriter struct{}

func newMockWriter() *mockWriter {
	return &mockWriter{}
}

func runRequest(b *testing.B, r *gin.Engine, method, path string) {
	// create fake request
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	w := newMockWriter()
	b.ReportAllocs()
	b.ResetTimer()

	//for i := 0; i < b.N; i++ {
	//	r.ServeHTTP(w, req)
	//}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.ServeHTTP(w, req)
		}
	})
}

func (m *mockWriter) Header() (h http.Header) {
	return map[string][]string{}
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockWriter) WriteHeader(int) {}

const enableDelay = false

func Empty(c *gin.Context) {
	if enableDelay {
		time.Sleep(50 * time.Microsecond)
	}
}

func BenchmarkFullServer(b *testing.B) {
	tracer.SetTracer(tc)
	e := gin.New()
	metrics.RegisterMetricsExport(e)
	c := gin_middles.DefaultChain()
	c[1] = gin_middles.GinLogger(func(option *gin_middles.LogOption) {
		option.Writer = &mockWriter{}
	})
	e.Use(c...)
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkLogServer(b *testing.B) {
	e := gin.New()
	e.Use(
		gin_middles.GinLogger(func(option *gin_middles.LogOption) {
			option.Writer = &mockWriter{}
		}),
	)
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkDefLogServer(b *testing.B) {
	e := gin.New()
	e.Use(
		gin.LoggerWithWriter(&mockWriter{}),
	)
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkRecoveryServer(b *testing.B) {
	e := gin.New()
	e.Use(
		gin_middles.Recovery(),
	)
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkEmptyServer(b *testing.B) {
	e := gin.New()
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkMetricsServer(b *testing.B) {
	e := gin.New()
	metrics.RegisterMetricsExport(e)
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkSentinelServer(b *testing.B) {
	e := gin.New()
	e.Use(sentinelgin.SentinelMiddleware())
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}

func BenchmarkTraceServer(b *testing.B) {
	tracer.SetTracer(tc)
	e := gin.New()
	e.Use(gin_middles.ExtractTrace())
	e.GET("/", Empty)
	runRequest(b, e, "GET", "")
}
