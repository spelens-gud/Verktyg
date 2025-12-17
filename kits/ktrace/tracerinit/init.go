package tracerinit

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	jaegerConfig "github.com/uber/jaeger-client-go/config"

	"github.com/spelens-gud/Verktyg.git/implements/otrace"

	"github.com/spelens-gud/Verktyg.git/kits/ktrace/tracer"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

var typeMapping = map[reflect.Type]reflect.Type{
	reflect.TypeOf(jaegerConfig.Configuration{}): reflect.TypeOf(otrace.JaegerConfig{}),
}

// 适配旧版本
// Deprecated: Use InitTracers
func InitTracer(configs ...interface{}) (tc itrace.Tracer, cf func(), err error) {
	traceConfigs := make([]itrace.TracerConfig, 0)
	for _, c := range configs {
		if tc, ok := c.(itrace.TracerConfig); ok {
			traceConfigs = append(traceConfigs, tc)
			continue
		}
		v := reflect.Indirect(reflect.ValueOf(c))
		if mt, ok := typeMapping[v.Type()]; ok {
			if tc, ok := v.Convert(mt).Interface().(itrace.TracerConfig); ok {
				traceConfigs = append(traceConfigs, tc)
			}
		}
	}
	return InitTracers(traceConfigs...)
}

func InitTracers(configs ...itrace.TracerConfig) (tc itrace.Tracer, cf func(), err error) {
	lg := logger.FromBackground()
	envTracer := strings.ToLower(os.Getenv(iconfig.EnvKeyTracerEnable))

	defer func() {
		err = nil
		if tc == nil {
			tc = itrace.NoopTracer
		}
		cf = func() { _ = tc.Close() }
		tracer.SetTracer(tc)
		lg.Infof("tracer init [ %s ]", tc)
	}()

	if len(envTracer) > 0 {
		lg.Infof("env tracer set [ %s ]", envTracer)
	}

	for _, c := range configs {
		if len(envTracer) > 0 {
			var match bool
			for _, n := range c.Names() {
				if n == envTracer {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if tc, err = c.NewTracer(); err == nil {
			return
		}
		lg.WithField("trace_config", fmt.Sprintf("%T", c)).Warnf("init tracer error: %v", err)
	}
	return
}
