package imetrics

import (
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	prom "github.com/prometheus/client_golang/prometheus"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type (
	ServerMetrics interface {
		prom.Collector
		// gin服务器中间件
		GinInterceptor() gin.HandlerFunc
		// 采集接口Handler
		Handler() http.HandlerFunc
		// 连接状态变更
		SetConnectionsState(conn net.Conn, state http.ConnState)
	}

	HttpServerMetricsReporter interface {
		Report(statusCode, code, retSize int)
	}

	HttpClientMetrics interface {
		prom.Collector

		NewReporter(peer, path, method string) HttpClientMetricsReporter

		ReportTcpDial(peer, address string, err error, duration time.Duration)
	}

	HttpClientMetricsReporter interface {
		Report(address string, code int, reqSize, retSize int64)
	}

	RpcServerMetrics interface {
		prom.Collector

		NewReporter(addr, serviceName, method, refererService string) RpcServerMetricsReporter
	}

	RpcServerMetricsReporter interface {
		Report(err error)
	}

	RpcClientMetrics interface {
		prom.Collector

		NewReporter(addr, serviceName, method string) (report func(err error))
	}

	GatewayDaemon interface {
		StartDaemon()
		Stop()
	}

	DBMetrics interface {
		prom.Collector
		Metrics(name, addr, command string, err error, duration time.Duration)
	}

	MqMetrics interface {
		prom.Collector
		MetricsProduce(addr, topic string, err error, duration time.Duration)
		MetricsConsume(addr, topic string, err error, duration time.Duration)
	}
)

const DefaultMetricsPath = "/metrics"

var registeredCollectors []prom.Collector

func init() {
	MustRegister(SYCoreVersion{})
}

func MustRegister(collectors ...prom.Collector) {
	for _, c := range collectors {
		if err := prom.Register(c); err != nil {
			pc, f, _, _ := runtime.Caller(1)
			logger.FromBackground().WithTag("METRICS_REGISTER_ERR").WithFields(map[string]interface{}{
				"file":     f,
				"function": runtime.FuncForPC(pc).Name(),
			}).Errorf("register prom error: %v", err)
		} else {
			registeredCollectors = append(registeredCollectors, c)
		}
	}
}

func GetRegisteredCollectors() []prom.Collector {
	return registeredCollectors
}
