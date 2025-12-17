package httpreq

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/spelens-gud/Verktyg.git/implements/promhttp"
	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg.git/interfaces/imetrics"
	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/ktrace/tracer"
)

type transporter struct {
	host       string
	status     int
	ctx        context.Context
	tStart     time.Time
	retryTimes int

	tDnsDone   time.Time
	tGotConn   time.Time
	tStartConn time.Time
	tStartResp time.Time
	tTlsStart  time.Time
	tTlsDone   time.Time

	tracingSpan     itrace.Span
	remoteAddr      net.Addr
	connAddr        string
	duration        time.Duration
	metricsReporter imetrics.HttpClientMetricsReporter

	option    *HttpDoOption
	request   *http.Request
	httpTrace *httptrace.ClientTrace

	metrics imetrics.HttpClientMetrics
}

var (
	tpPool = sync.Pool{New: func() interface{} {
		return &transporter{}
	}}

	tracePool = sync.Pool{New: func() interface{} {
		return &httptrace.ClientTrace{}
	}}
)

func (tp *transporter) GotConn(info httptrace.GotConnInfo) {
	tp.remoteAddr = info.Conn.RemoteAddr()
	tp.tGotConn = time.Now()
}

func (tp *transporter) TlsStart() {
	tp.tTlsStart = time.Now()
}

func (tp *transporter) TlsDone(tls.ConnectionState, error) {
	tp.tTlsDone = time.Now()
}

func (tp *transporter) GetConn(addr string) {
	tp.connAddr = addr
	tp.tStartConn = time.Now()
}

func (tp *transporter) ConnectDone(_, addr string, err error) {
	if tp.tStartConn.IsZero() {
		return
	}
	tp.metrics.ReportTcpDial(tp.host, addr, err, time.Since(tp.tStartConn))
}

func (tp *transporter) DNSDone(_ httptrace.DNSDoneInfo) {
	tp.tDnsDone = time.Now()
}

func (tp *transporter) GotFirstResponseByte() {
	tp.tStartResp = time.Now()
}

func (tp *transporter) init(request *http.Request) (newRequest *http.Request) {
	retryTimes, _ := contextKeyRetriedTime.FromHttpRequest(request).(int)

	*tp = transporter{
		status:     -1,
		ctx:        request.Context(),
		option:     getRequestDoOption(request),
		tStart:     time.Now(),
		host:       request.Host,
		retryTimes: retryTimes,
		request:    request,
		metrics:    promhttp.DefaultClientMetrics,
	}

	tp.metricsReporter = tp.metrics.NewReporter(request.URL.Host, request.URL.Path, request.Method)

	if !tp.option.DisableHttpTrace {
		// http内部追踪信息
		tp.request = tp.initHttpTrace(tp.request)
	}

	if tp.option.DisableTrace {
		// 注入空链路追踪
		tp.tracingSpan = itrace.NoopSpan
	} else {
		// 链路追踪注入
		tp.tracingSpan = tracer.InjectHttp(tp.request)
	}

	return tp.request
}

func (tp *transporter) log(retSize int64, httpError error) {
	// logger
	var (
		fields = tp.initLogFields(retSize)
		logger = tp.option.logger(tp.ctx).WithTag("HTTP_CLIENT").WithFields(fields)
	)

	// 打印日志
	if httpError != nil {
		// 整理请求日志信息
		logger.WithTag("HTTP_CLIENT_ERR").Errorf("http client request error: %v", httpError)
	} else if tp.status >= 500 {
		// 返回状态5XX
		logger.Error("http client request finished")
	} else if !tp.option.DisableLog {
		// 整理请求日志信息
		logger.Info("http client request finished")
	}
}

func (tp *transporter) reportMetrics(retSize int64) {
	tp.metricsReporter.Report(tp.request.URL.Host, tp.status, tp.request.ContentLength, retSize)
}

func (tp *transporter) reportTrace(err error) {
	if tp.retryTimes > 0 {
		tp.tracingSpan.Tag("http.retry_times", strconv.Itoa(tp.retryTimes))
	}
	// tracing tag
	itrace.SetHttpStatusTag(tp.tracingSpan, tp.status, err)
	if tp.remoteAddr != nil {
		itrace.SetNetPeerTag(tp.tracingSpan, tp.remoteAddr)
	}
	// end
	tp.tracingSpan.Finish()
}

func (tp *transporter) initHttpTrace(r *http.Request) *http.Request {
	tp.httpTrace = tracePool.Get().(*httptrace.ClientTrace)
	*tp.httpTrace = httptrace.ClientTrace{
		GetConn:              tp.GetConn,
		GotConn:              tp.GotConn,
		GotFirstResponseByte: tp.GotFirstResponseByte,
		DNSDone:              tp.DNSDone,
		ConnectDone:          tp.ConnectDone,
		TLSHandshakeStart:    tp.TlsStart,
		TLSHandshakeDone:     tp.TlsDone,
	}
	return r.WithContext(httptrace.WithClientTrace(r.Context(), tp.httpTrace))
}

func (tp *transporter) initLogFields(retSize int64) (fields map[string]interface{}) {
	request := tp.request
	fields = map[string]interface{}{
		"host":         request.URL.Host,
		"path":         request.URL.Path,
		"url":          request.URL.String(),
		"method":       request.Method,
		"request_size": request.ContentLength,
		"status":       tp.status,
		"duration":     tp.duration.String(),
		"return_size":  retSize,
		"cost":         tp.duration.Seconds(),
	}

	if tp.remoteAddr != nil {
		fields["remote_addr"] = tp.remoteAddr.String()
	} else {
		fields["remote_addr"] = tp.connAddr
	}

	// 重试次数
	if tp.retryTimes > 0 {
		fields["retry_times"] = tp.retryTimes
	}

	// 连接建立信息
	if !tp.tGotConn.IsZero() {
		fields["tcp_conn_duration"] = tp.tGotConn.Sub(tp.tStartConn).String()
	}
	// 等待返回时间
	if !tp.tStartResp.IsZero() {
		fields["resp_wait_duration"] = tp.tStartResp.Sub(tp.tGotConn).String()
	}
	// DNS寻址时间
	if !tp.tDnsDone.IsZero() {
		fields["dns_resolve_duration"] = tp.tDnsDone.Sub(tp.tStart).String()
	}
	// TLS握手时间
	if !tp.tTlsDone.IsZero() && !tp.tTlsStart.IsZero() {
		fields["tls_handshake_duration"] = tp.tTlsDone.Sub(tp.tTlsStart).String()
	}
	return fields
}

func RoundTrip(doRoundTrip func(*http.Request) (*http.Response, error), request *http.Request) (resp *http.Response, err error) {
	if contextKeyInTransport.FromHttpRequest(request) != nil {
		return doRoundTrip(request)
	}

	// 标记已经进入拦截器 避免循环处理
	request = contextKeyInTransport.WithHttpRequest(request, struct{}{})

	tp := tpPool.Get().(*transporter)
	request = tp.init(request)

	defer func() {
		var (
			// 返回体字节数
			retSize int64
			// 错误处理
			httpError = err
		)

		// 重定向拦截 不作为错误处理
		// 如果client有修改CheckRedirect会有可能导致 resp和err同时不为nil
		if urlErr, _ := httpError.(*url.Error); urlErr != nil {
			if _, is := urlErr.Err.(ihttp.IRedirectError); is {
				httpError = nil
			}
		}

		// 请求时间
		tp.duration = time.Since(tp.tStart)

		// 返回体字节数
		if resp != nil {
			tp.status = resp.StatusCode

			if resp.ContentLength > 0 {
				retSize = resp.ContentLength
			} else {
				retSize, _ = strconv.ParseInt(resp.Header.Get(ihttp.HeaderContentLength), 10, 64)
			}
		}

		// 监控
		tp.reportMetrics(retSize)
		// 日志
		tp.log(retSize, httpError)
		// 链路追踪
		tp.reportTrace(httpError)

		// 对象池复用
		tracePool.Put(tp.httpTrace)
		tpPool.Put(tp)
	}()

	return doRoundTrip(request)
}

func WrapRoundTripper(r http.RoundTripper, option *HttpDoOption) http.RoundTripper {
	return ihttp.WrapRoundTripper(r, func(doRoundTrip func(*http.Request) (*http.Response, error), request *http.Request) (resp *http.Response, err error) {
		if option != nil {
			request = contextKeyHttpOption.WithHttpRequest(request, option)
		}
		return RoundTrip(doRoundTrip, request)
	})
}
