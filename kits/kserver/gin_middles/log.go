package gin_middles

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"git.bestfulfill.tech/devops/go-core/kits/kgo"

	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/kenv/envflag"
	"git.bestfulfill.tech/devops/go-core/kits/kgo/buffpool"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
	"git.bestfulfill.tech/devops/go-core/kits/kserver/log"
)

var (
	enableOmitOkLog   = envflag.BoolOnceFrom(iconfig.EnvKeyServerEnableOKLogOmit)
	disableLogReqBody = envflag.BoolOnceFrom(iconfig.EnvKeyLogDisableRequestBody)

	lineBreak = []byte{'\n'}
)

type LogOption struct {
	ClientIpFunc      func(c *gin.Context) string
	Ignore2XXStatus   bool
	IgnoreStatus      func(status int) bool
	DisableLogReqBody bool
	IgnoreRoutes      []string
	Writer            io.Writer
}

func GinLogger(opts ...func(*LogOption)) gin.HandlerFunc {
	opt := &LogOption{
		ClientIpFunc:      defaultClientIPFunc,
		Ignore2XXStatus:   enableOmitOkLog(),
		DisableLogReqBody: disableLogReqBody(),
		Writer:            logger.Writer(),
	}

	for _, o := range opts {
		o(opt)
	}

	for i := range opt.IgnoreRoutes {
		opt.IgnoreRoutes[i] = strings.TrimSuffix(opt.IgnoreRoutes[i], "/") + "/"
	}

	writer := opt.Writer

	return func(c *gin.Context) {
		for _, p := range opt.IgnoreRoutes {
			if strings.HasPrefix(c.FullPath(), p) || c.FullPath() == p[:len(p)-1] {
				return
			}
		}

		var (
			start   = time.Now()
			request = c.Request

			cookieData string
			copyBuff   *bytes.Buffer
		)

		if !opt.DisableLogReqBody && request.Body != nil {
			copyBuff = buffpool.GetBytesBuffer()
			defer buffpool.PutBytesBuffer(copyBuff)

			request.Body = ioutil.NopCloser(io.TeeReader(request.Body, copyBuff))
			cookieData = c.Request.Header.Get("Cookie")
		}

		defer func() {
			status := c.Writer.Status()

			if opt.Ignore2XXStatus && status/100 == 2 {
				return
			}

			if opt.IgnoreStatus != nil && opt.IgnoreStatus(status) {
				return
			}

			info := log.HttpLogInfoFromRequest(c.Request, start, status, opt.ClientIpFunc(c))

			if copyBuff != nil && copyBuff.Len() > 0 {
				info.PostData = kgo.UnsafeBytes2string(copyBuff.Bytes())
			}

			if len(cookieData) > 0 {
				info.CookieData = cookieData
			}

			_, _ = writer.Write(info.Marshal())
			_, _ = writer.Write(lineBreak)

			log.PutServerLoggerInfo(info)
		}()

		c.Next()
	}
}
