package gin_middles

import (
	"bytes"
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

type RecoveryOption struct {
	PanicStatus int
}

func Recovery(opts ...func(*RecoveryOption)) gin.HandlerFunc {
	opt := &RecoveryOption{
		PanicStatus: defaultPanicStatus,
	}
	for _, o := range opts {
		o(opt)
	}
	return func(c *gin.Context) {
		defer opt.recovery(c)
		c.Next()
	}
}

func (opt *RecoveryOption) recovery(c *gin.Context) {
	e := recover()
	if e == nil {
		return
	}

	var (
		brokenPipe bool
		lg         = logger.FromContext(c.Request.Context())
	)

	if ne, ok := e.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			errMsg := strings.ToLower(se.Error())
			brokenPipe = strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset by peer")
		}
	}

	// tcp连接导致的panic
	if brokenPipe {
		if e, ok := e.(error); ok {
			_ = c.Error(e)
		}
		c.Abort()
		lg.WithTag("SERVER_BROKEN_PIPE").WithField("remote_addr", c.Request.RemoteAddr).Errorf("%v", e)
		return
	}

	var (
		stack          = stack(3)
		httpRequest, _ = httputil.DumpRequest(c.Request, false)
		headers        = strings.Split(string(httpRequest), "\r\n")
	)

	for idx, header := range headers {
		current := strings.Split(header, ":")
		if current[0] == "Authorization" {
			headers[idx] = current[0] + ": *"
		}
	}

	// 打印日志
	if !c.Writer.Written() {
		c.AbortWithStatus(opt.PanicStatus)
	}

	fmt.Printf("[Recovery] %s panic recovered:\n%s\n%s\n%s", timeFormat(time.Now()), strings.Join(headers, "\r\n"), e, stack)

	lg.WithTag("SERVER_PANIC").WithField("stacks", string(stack)).Errorf("[Recovery] panic recovered: %v", e)
}

func stack(skip int) (b []byte) {
	var (
		buf      = new(bytes.Buffer)
		lines    [][]byte
		lastFile string
	)
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexe
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.ReplaceAll(name, centerDot, dot)
	return name
}

func timeFormat(t time.Time) string {
	var timeString = t.Format("2006/01/02 15:04:05")
	return timeString
}
