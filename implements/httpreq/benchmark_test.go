package httpreq

import (
	"net/http"
	"strings"
	"testing"

	"github.com/spelens-gud/Verktyg/interfaces/ierror"

	"github.com/eapache/go-resiliency/breaker"

	"github.com/spelens-gud/Verktyg/kits/kerror/errorx"

	"github.com/spelens-gud/Verktyg/interfaces/ilog"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

var handlerFunc = func() {
	_ = http.ListenAndServe(":80", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
}

var makeReq = func() *http.Request {
	req, _ := http.NewRequest("POST", "http://localhost:80", strings.NewReader("ffffff"))
	return req
}

func TestErr(t *testing.T) {
	t.Log(IsBreakOpenErr(errorx.Wrap(breaker.ErrBreakerOpen, ierror.ResourceExhausted)))
}

func BenchmarkDoRequest(b *testing.B) {
	logger.SetLevel(ilog.Error)
	opt := WithHttpDoOption(&HttpDoOption{
		DisableHttpTrace: true,
		DisableTrace:     true,
		DisableLog:       true,
		MaxRetryTimes:    0,
		RetryCheck:       nil,
	})
	_ = opt
	go handlerFunc()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := makeReq()
			opt(req)
			ret, err := doRequest(http.DefaultClient, req)
			if err != nil {
				b.Fatal(err)
			}
			_ = ret
		}
	})
}

func BenchmarkDoRequestRaw(b *testing.B) {
	logger.SetLevel(ilog.Error)
	go handlerFunc()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ret, err := http.DefaultClient.Do(makeReq())
			if err != nil {
				b.Fatal(err)
			}
			_ = ret
		}
	})
}
