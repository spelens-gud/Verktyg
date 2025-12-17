package promrpc

import (
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/imetrics"
)

type ClitMetrics struct {
	*imetrics.MetricsGroup
}

func NewCliMetrics() imetrics.RpcClientMetrics {
	group := imetrics.NewMetricsGroup(imetrics.NamespaceRpc, imetrics.SubsystemClient, nil)

	group.NewHistogram(imetrics.NameHandlingSeconds, "Histogram of response latency (seconds) of Http that had been application-level handled by the client.",
		"rpc_method", "rpc_service", "rpc_addr", "rpc_error",
	)

	m := &ClitMetrics{group}
	imetrics.MustRegister(m)
	return m
}

func (m *ClitMetrics) NewReporter(addr, serviceName, method string) (report func(err error)) {
	t := time.Now()
	return func(err error) {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		m.Add(imetrics.NameHandlingSeconds, time.Since(t).Seconds(), method, serviceName, addr, errMsg)
	}
}
