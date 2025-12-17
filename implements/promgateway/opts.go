package promgateway

import (
	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
)

func WithIntervalSeconds(s int) Options {
	if s == 0 {
		s = 5
	}
	if s < 0 {
		panic("invalid interval")
	}
	return func(option *GatewayConfig) {
		option.IntervalSeconds = s
	}
}

func WithGroupingName(name string) Options {
	return func(option *GatewayConfig) {
		option.GroupingName = name
	}
}

func WithGroupingKV(f func() (key, value string)) Options {
	return func(option *GatewayConfig) {
		option.GroupingKV = append(option.GroupingKV, f)
	}
}

func WithHttpReqOptions(reqOpt ...ihttp.ReqOpt) Options {
	return func(option *GatewayConfig) {
		option.ReqOpt = append(option.ReqOpt, reqOpt...)
	}
}
