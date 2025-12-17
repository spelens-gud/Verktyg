package httpreq

import (
	"context"
	"net/http"
	"sync"

	"github.com/spelens-gud/Verktyg/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg/interfaces/ilog"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

type HttpDoOption struct {
	DisableTrace      bool
	DisableLog        bool
	DisableHttpTrace  bool
	DisableMetaHeader bool
	MaxRetryTimes     int
	RetryCheck        func(ret *http.Response, retryTimes int) (can bool)
	Logger            ilog.Logger
}

// 默认执行配置
var (
	defaultDoOption = &HttpDoOption{
		MaxRetryTimes: DefaultMaxRetryTimes,
	}
	mu sync.Mutex
)

func SetDefaultDoOption(option *HttpDoOption) {
	mu.Lock()
	defaultDoOption = option
	mu.Unlock()
}

func (o *HttpDoOption) logger(ctx context.Context) (lg ilog.Logger) {
	if o.Logger != nil {
		lg = o.Logger
	} else {
		lg = logger.FromContext(ctx)
	}
	return lg
}

func WithHttpDoOption(option *HttpDoOption) ihttp.ReqOpt {
	return func(r *http.Request) {
		*r = *contextKeyHttpOption.WithHttpRequest(r, option)
	}
}

func WithDoOptionContext(ctx context.Context, option *HttpDoOption) context.Context {
	return contextKeyHttpOption.WithValue(ctx, option)
}

func getRequestDoOption(req *http.Request) *HttpDoOption {
	if o, _ := contextKeyHttpOption.FromHttpRequest(req).(*HttpDoOption); o != nil {
		return o
	}
	return defaultDoOption
}
