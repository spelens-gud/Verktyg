package promhttp

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"git.bestfulfill.tech/devops/go-core/interfaces/imetrics"
	"git.bestfulfill.tech/devops/go-core/kits/kcontext"
)

type serverMetrics struct {
	*imetrics.MetricsGroup
}

var DefaultServerMetrics = NewServerMetrics()

func NewServerMetrics() imetrics.ServerMetrics {
	group := imetrics.NewMetricsGroup(imetrics.NamespaceHttp, imetrics.SubsystemServer, nil)

	// 入口请求body bytes计数
	group.NewCounter("io_bytes", "Total number of Http IO bytes handle on the server.",
		"http_host", "http_path", "http_method", "http_code", "service_code", "referer_service", "type",
	)

	// 入口请求body bytes计数
	group.NewCounter("connections_state", "Total number of connections of state  on the server.",
		"state", "http_addr",
	)

	// 服务时延计数
	group.NewHistogram(imetrics.NameHandlingSeconds, "Histogram of response latency (seconds) of Http that had been application-level handled by the server.",
		"http_host", "http_path", "http_method", "http_code", "service_code", "referer_service",
	)

	ret := &serverMetrics{group}
	imetrics.MustRegister(ret)
	return ret
}

type ServerReporter struct {
	*serverMetrics
	host, fullPath, method, refererService string
	reqSize                                int64
	tStart                                 time.Time
}

func (reporter *ServerReporter) Report(statusCode, code, retSize int) {
	if retSize < 0 {
		retSize = 0
	}

	labels := map[string]string{
		"http_host":       reporter.host,
		"http_path":       reporter.fullPath,
		"http_method":     reporter.method,
		"referer_service": reporter.refererService,
		"http_code":       strconv.Itoa(statusCode),
		"service_code":    strconv.Itoa(code),
	}

	// 请求时延
	reporter.AddX(imetrics.NameHandlingSeconds, time.Since(reporter.tStart).Seconds(), labels)
	// 请求body bytes计数
	reporter.AddX("io_bytes", float64(reporter.reqSize), labels, map[string]string{"type": "request"})
	// 返回body bytes计数
	reporter.AddX("io_bytes", float64(retSize), labels, map[string]string{"type": "return"})

	reporterPool.Put(reporter)
}

func (reporter *ServerReporter) ReportGinContext(c *gin.Context) {
	var (
		retSize = 0
		status  = -1
	)
	if c.Writer.Status() > 0 {
		status = c.Writer.Status()
		retSize = c.Writer.Size()
	}
	reporter.Report(status, kcontext.GetServiceCode(c.Request.Context()), retSize)
}

var reporterPool = sync.Pool{New: func() interface{} {
	return new(ServerReporter)
}}

func (m *serverMetrics) NewReporter(host, fullPath, method, refererService string, reqSize int64) *ServerReporter {
	if reqSize < 0 {
		reqSize = 0
	}

	reporter := reporterPool.Get().(*ServerReporter)
	*reporter = ServerReporter{
		serverMetrics:  m,
		host:           host,
		fullPath:       fullPath,
		method:         method,
		refererService: refererService,
		reqSize:        reqSize,
		tStart:         time.Now(),
	}
	return reporter
}

func (m *serverMetrics) GinInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			host           = c.Request.Host
			method         = c.Request.Method
			refererService = c.Request.Header.Get(kcontext.HeaderRefererService)
		)

		if len(host) == 0 {
			host = c.Request.URL.Host
		}

		defer m.NewReporter(host, c.FullPath(), method, refererService, c.Request.ContentLength).ReportGinContext(c)

		c.Next()
	}
}

func (m *serverMetrics) SetConnectionsState(conn net.Conn, state http.ConnState) {
	m.Add("connections_state", 1, state.String(), conn.LocalAddr().String())
}

func (m *serverMetrics) Handler() http.HandlerFunc {
	return promhttp.Handler().ServeHTTP
}
