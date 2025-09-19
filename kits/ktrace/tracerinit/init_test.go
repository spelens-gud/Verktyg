package tracerinit

import (
	"os"
	"testing"

	jaegerConfig "github.com/uber/jaeger-client-go/config"

	"git.bestfulfill.tech/devops/go-core/implements/otrace"

	"git.bestfulfill.tech/devops/go-core/implements/skytrace"
	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
)

func TestInit(t *testing.T) {
	_ = os.Setenv(iconfig.EnvKeyTracerEnable, "jaeger")
	tr, cf, err := InitTracers(skytrace.GRPCReportTracerConfig{
		Service:      "",
		Addr:         "",
		Auth:         "",
		CheckSeconds: 0,
		SampleRate:   0,
	}, otrace.JaegerConfig{
		ServiceName:         "",
		Disabled:            false,
		RPCMetrics:          false,
		Tags:                nil,
		Sampler:             nil,
		Reporter:            nil,
		Headers:             nil,
		BaggageRestrictions: nil,
		Throttler:           nil,
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer cf()
	t.Log(tr)
}

func TestInit2(t *testing.T) {
	tr, cf, err := InitTracer(skytrace.GRPCReportTracerConfig{
		Service:      "",
		Addr:         "",
		Auth:         "",
		CheckSeconds: 0,
		SampleRate:   0,
	}, jaegerConfig.Configuration{
		ServiceName:         "xxx",
		Disabled:            false,
		RPCMetrics:          false,
		Tags:                nil,
		Sampler:             nil,
		Reporter:            nil,
		Headers:             nil,
		BaggageRestrictions: nil,
		Throttler:           nil,
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer cf()
	t.Log(tr)
}
