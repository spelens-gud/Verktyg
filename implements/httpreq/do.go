package httpreq

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg/kits/kcontext"
	"github.com/spelens-gud/Verktyg/kits/kerror/errorx"
	"github.com/spelens-gud/Verktyg/kits/kgo/buffpool"
	"github.com/spelens-gud/Verktyg/version"
)

const (
	DefaultMaxRetryTimes = 2
	MaxBackoffTime       = time.Second * 30

	HeaderGoCoreVersion = "X-Go-Core-Version"
)

func do(cli *http.Client, request *http.Request, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	response, err := doRequest(cli, request, opts...)
	if err != nil {
		return
	}
	// 请求完成
	resp, err = NewResp(response)
	return
}

func DoRequest(cli *http.Client, request *http.Request, opts ...ihttp.ReqOpt) (ret *http.Response, err error) {
	return doRequest(cli, request, opts...)
}

func setMetaHeader(request *http.Request) {
	if len(request.Header.Get(HeaderGoCoreVersion)) == 0 {
		// 调用端版本信息
		request.Header.Set(HeaderGoCoreVersion, version.GetVersion())
	}

	if len(request.Header.Get(kcontext.HeaderRefererService)) == 0 {
		// 调用服务来源标识
		request.Header.Set(kcontext.HeaderRefererService, iconfig.GetApplicationName())
	}
}

func doRequest(cli *http.Client, request *http.Request, opts ...ihttp.ReqOpt) (ret *http.Response, err error) {
	for _, o := range opts {
		o(request)
	}

	// 自定义调用细节配置
	doOpt := getRequestDoOption(request)

	// sy-core头部
	if !doOpt.DisableMetaHeader {
		setMetaHeader(request)
	}

	// 如果deadline exceeded 转化为 code error
	defer func() {
		if errors.Is(err, context.DeadlineExceeded) {
			err = errorx.ErrDeadlineExceeded("do http request context error: " + err.Error())
		} else if errors.Is(err, context.Canceled) {
			err = errorx.ErrCancelled("do http request context error: " + err.Error())
		}
	}()

	// 不需要重试
	if doOpt.MaxRetryTimes == 0 {
		return RoundTrip(cli.Do, request)
	}

	// 需要重试的情况
	var (
		retriedTimes = -1
		backoffTimer = time.NewTimer(defaultBackoffUnit)
		retryCheck   = doOpt.RetryCheck
	)

	// 重试检查
	if retryCheck == nil {
		retryCheck = doOpt.defaultRetryCheck
	}

	if request.Body != nil && request.GetBody == nil {
		// 自动备份数据处理预备
		newBuffer := buffpool.GetBytesBuffer()
		defer buffpool.PutBytesBuffer(newBuffer)

		var (
			copyReader bytes.Reader
			readerPtr  *bytes.Reader

			readerCarrier = &struct{ io.Reader }{io.TeeReader(request.Body, newBuffer)}
		)

		request.Body = io.NopCloser(readerCarrier)
		request.GetBody = func() (io.ReadCloser, error) {
			if readerPtr == nil {
				readerPtr = bytes.NewReader(newBuffer.Bytes())
				copyReader = *readerPtr
				readerCarrier.Reader = readerPtr
			} else {
				*readerPtr = copyReader
			}
			return request.Body, nil
		}
	}

	// 正式开始请求
Retry:
	retriedTimes++

	if retriedTimes > 0 && request.Body != nil {
		if request.Body, err = request.GetBody(); err != nil {
			return
		}
	}

	request = contextKeyRetriedTime.WithHttpRequest(request, retriedTimes)

	// 请求处理
	if ret, err = RoundTrip(cli.Do, request); err != nil {
		return
	}

	status := -1

	if ret != nil {
		status = ret.StatusCode
		// 2XX即请求成功
		if (status / 100) == 2 {
			return
		}

		// 检查重试策略
		if !retryCheck(ret, retriedTimes) {
			return
		}

		// 关闭连接 等待退避重试
		_ = ret.Body.Close()
	}

	var (
		ctx         = request.Context()
		backoffTime = getBackoffTime(retriedTimes, status)
	)

	// 重试日志
	if !doOpt.DisableLog {
		doOpt.logger(ctx).Warnf("retry [ %s ] after back off [ %s ]", request.URL.String(), backoffTime.String())
	}

	backoffTimer.Reset(backoffTime)

	select {
	// 上层ctx取消
	case <-ctx.Done():
		backoffTimer.Stop()
		err = ctx.Err()
		return
	case <-backoffTimer.C:
		backoffTimer.Stop()
		goto Retry
	}
}

// 检查重试
func (o HttpDoOption) defaultRetryCheck(ret *http.Response, retryTimes int) (can bool) {
	// 重试次数检查
	if retryTimes >= o.MaxRetryTimes {
		return false
	}

	// 501
	if ret.StatusCode == http.StatusNotImplemented {
		return false
	}

	// 429
	if ret.StatusCode == http.StatusTooManyRequests {
		return true
	}

	// 5XX
	return ret.StatusCode >= 500
}

const defaultBackoffUnit = time.Millisecond * 5

func getBackoffTime(retryTimes, statusCode int) time.Duration {
	// 504 429等压力超时 需要延长退避重试时间
	if statusCode == http.StatusTooManyRequests || statusCode == http.StatusGatewayTimeout {
		retryTimes += 1
	}
	x := 1 << retryTimes
	// 指数退避
	t := defaultBackoffUnit * time.Duration(x)
	if t > MaxBackoffTime {
		t = MaxBackoffTime
	}
	return t
}
