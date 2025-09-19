package otrace

import (
	"context"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"

	"git.bestfulfill.tech/devops/go-core/interfaces/itrace"
)

func TestExtractRequestID(t *testing.T) {
	s, _ := jaeger.NewProbabilisticSampler(0.01)
	tracer, cf, err := (jaegerConfig.Configuration{
		ServiceName: "test",
		RPCMetrics:  true,
		Reporter: &jaegerConfig.ReporterConfig{
			LocalAgentHostPort: "127.0.0.1:6831",
		},
	}).NewTracer(jaegerConfig.Sampler(s))
	if err != nil {
		t.Fatal(err)
	}
	// nolint
	defer cf.Close()
	opentracing.SetGlobalTracer(tracer)

	tr := &Tracer{}
	ctx := context.Background()
	header := map[string][]string{}
	ctx = itrace.WithContext(ctx, "3e9941813c9b307b")
	sp := tr.InjectMetadata(ctx, "test-in", "no", header)
	defer sp.Finish()

	for i := 0; i < 1000; i++ {
		sp2, _ := tr.ExtractFromMetadata(ctx, "test", header)
		ds := sp2.(*Span).Span.(*jaeger.Span).SpanContext().TraceID().String()
		t.Log(ds)
		sp2.Finish()

	}

	defer time.Sleep(time.Second)
}
