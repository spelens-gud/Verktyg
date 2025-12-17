package otrace

import (
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

type JaegerConfig jaegerConfig.Configuration

func (config JaegerConfig) Names() []string {
	return []string{"jaeger", "opentracing"}
}

func (config *JaegerConfig) init() {
	if config.Sampler == nil || len(config.Sampler.Type) == 0 {
		config.Sampler = &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 0.001,
		}
	}
}

func (config JaegerConfig) NewTracer() (t itrace.Tracer, err error) {
	if len(config.ServiceName) == 0 {
		config.ServiceName = iconfig.GetApplicationName()
	}
	config.init()
	return NewOpenTracingTracer(jaegerConfig.Configuration(config).NewTracer(jaegerConfig.Logger(logger.FromBackground().WithTag("JAEGER"))))
}
