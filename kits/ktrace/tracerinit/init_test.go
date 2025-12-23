package tracerinit

import (
	"os"
	"testing"

	jaegerConfig "github.com/uber/jaeger-client-go/config"

	"github.com/spelens-gud/Verktyg/implements/otrace"

	"github.com/spelens-gud/Verktyg/implements/skytrace"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
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
