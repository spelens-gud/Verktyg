package gin_middles

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/kcontext"
	"github.com/spelens-gud/Verktyg.git/kits/ktrace/tracer"
)

var defaultHeaderRequestIDKeys = []string{itrace.HeaderXRequestID, "Request-Id", itrace.HeaderXTraceID}

type TraceOption struct {
	ClientIpFunc                func(c *gin.Context) string
	HeaderRequestIDKeys         []string
	ResponseHeaderRequestIDKeys []string
	PanicStatus                 int
}

func ExtractTrace(opts ...func(*TraceOption)) gin.HandlerFunc {
	opt := &TraceOption{
		ClientIpFunc:                defaultClientIPFunc,
		HeaderRequestIDKeys:         defaultHeaderRequestIDKeys,
		ResponseHeaderRequestIDKeys: []string{itrace.HeaderXRequestID},
		PanicStatus:                 defaultPanicStatus,
	}

	for _, o := range opts {
		o(opt)
	}

	return func(c *gin.Context) {
		// 链路追踪导出上游
		sp, ctx := tracer.ExtractHttp(c.Request, opt.HeaderRequestIDKeys...)
		// IP地址
		itrace.SetPeerIPTag(sp, net.ParseIP(opt.ClientIpFunc(c)))

		defer func() {
			var (
				e        = recover()
				status   = c.Writer.Status()
				afterCtx = c.Request.Context()

				err         error
				serviceCode int
			)

			if e != nil {
				// 从panic中恢复
				err = fmt.Errorf("%v", e)
				status = opt.PanicStatus
			} else {
				// 已知的业务错误
				serviceCode = kcontext.GetServiceCode(afterCtx)
				err = kcontext.GetRequestError(afterCtx)
				itrace.SetServiceCodeTag(sp, serviceCode)

				if status >= 500 && err == nil {
					err = fmt.Errorf("service error: %d", serviceCode)
				}
			}

			itrace.SetHttpStatusTag(sp, status, err)

			sp.Finish()

			// 抛出panic给recovery处理
			if e != nil {
				panic(e)
			}
		}()

		kcontext.SetRequestContext(c.Request, ctx)
		itrace.SetMetadataRequestID(itrace.FromContext(ctx), c.Writer.Header(), opt.ResponseHeaderRequestIDKeys...)
		c.Next()
	}
}
