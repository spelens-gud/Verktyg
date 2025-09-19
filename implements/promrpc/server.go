package promrpc

import (
	"time"

	"git.bestfulfill.tech/devops/go-core/interfaces/imetrics"
)

type ServerMetrics struct {
	*imetrics.MetricsGroup
}

func NewServerMetrics() imetrics.RpcServerMetrics {
	group := imetrics.NewMetricsGroup(imetrics.NamespaceRpc, imetrics.SubsystemServer, nil)

	group.NewHistogram(imetrics.NameHandlingSeconds, "Histogram of response latency (seconds) of Http that had been application-level handled by the server.",
		"rpc_method", "rpc_service", "rpc_addr", "referer_service", "rpc_error",
	)

	m := &ServerMetrics{group}
	imetrics.MustRegister(m)
	return m
}

type ServerReporter struct {
	*ServerMetrics
	method         string
	serviceName    string
	addr           string
	refererService string
	tStart         time.Time
}

func (u ServerReporter) Report(err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	u.Add(imetrics.NameHandlingSeconds, time.Since(u.tStart).Seconds(), u.method, u.serviceName, u.addr, u.refererService, errMsg)
}

func (m *ServerMetrics) NewReporter(addr, serviceName, method, refererService string) imetrics.RpcServerMetricsReporter {
	return ServerReporter{
		ServerMetrics:  m,
		method:         method,
		serviceName:    serviceName,
		addr:           addr,
		refererService: refererService,
		tStart:         time.Now(),
	}
}
