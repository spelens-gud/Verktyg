package promhttp

import (
	"strconv"
	"time"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/interfaces/imetrics"
	"git.bestfulfill.tech/devops/go-core/kits/kenv/envflag"
)

type ClitMetrics struct {
	*imetrics.MetricsGroup
}

var DefaultClientMetrics = NewCliMetrics()

func NewCliMetrics() imetrics.HttpClientMetrics {
	group := imetrics.NewMetricsGroup(imetrics.NamespaceHttp, imetrics.SubsystemClient, nil)

	group.NewCounter("io_bytes", "Total number of Http IO bytes handle on the client.",
		"http_host", "http_peer", "http_path", "http_method", "http_code", "type")

	group.NewHistogram(imetrics.NameHandlingSeconds, "Histogram of response latency (seconds) of Http that had been application-level handled by the client.",
		"http_host", "http_peer", "http_path", "http_method", "http_code")

	group.NewHistogram("dial_seconds", "Histogram of tcp connect latency (seconds) of Http by the client.",
		"http_host", "http_peer", "error")

	m := &ClitMetrics{group}
	imetrics.MustRegister(m)
	return m
}

var enableClientPathMetrics = envflag.BoolOnceFrom(iconfig.EnvKeyRuntimeMetricsEnableClientPath)

type UnwrapError interface {
	Unwrap() error
}

func (m *ClitMetrics) ReportTcpDial(peer, address string, err error, duration time.Duration) {
	msg := ""
	if u, ok := err.(UnwrapError); ok {
		err = u.Unwrap()
	}
	if err != nil {
		msg = err.Error()
	}
	m.Add("dial_seconds", duration.Seconds(), peer, address, msg)
}

type CliReporter struct {
	*ClitMetrics
	host   string
	path   string
	method string
	tStart time.Time
}

func (c CliReporter) Report(peer string, code int, reqSize, retSize int64) {
	if reqSize < 0 {
		reqSize = 0
	}

	if retSize < 0 {
		retSize = 0
	}

	labels := map[string]string{
		"http_host":   c.host,
		"http_peer":   peer,
		"http_path":   c.path,
		"http_method": c.method,
		"http_code":   strconv.Itoa(code),
	}

	c.AddX(imetrics.NameHandlingSeconds, time.Since(c.tStart).Seconds(), labels)
	c.AddX("io_bytes", float64(reqSize), labels, map[string]string{"type": "request"})
	c.AddX("io_bytes", float64(retSize), labels, map[string]string{"type": "return"})
}

var _ imetrics.HttpClientMetricsReporter = CliReporter{}

func (m *ClitMetrics) NewReporter(peer, path, method string) imetrics.HttpClientMetricsReporter {
	if !enableClientPathMetrics() {
		path = ""
	}
	return CliReporter{
		ClitMetrics: m,
		host:        peer,
		path:        path,
		method:      method,
		tStart:      time.Now(),
	}
}
