package log

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/kcontext"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return &ServerLoggerInfo{}
	},
}

func PutServerLoggerInfo(info *ServerLoggerInfo) {
	pool.Put(info)
}

func GetServerLoggerInfo() *ServerLoggerInfo {
	return pool.Get().(*ServerLoggerInfo)
}

func HttpLogInfoFromRequest(request *http.Request, start time.Time, status int, clientIp string) *ServerLoggerInfo {
	var (
		header   = request.Header
		now      = time.Now()
		cost     = now.Sub(start)
		reqPath  = request.URL.Path
		rawQuery = request.URL.RawQuery
		size     = request.ContentLength
		info     = GetServerLoggerInfo()
	)

	info.Reset()

	if rawQuery != "" {
		reqPath = reqPath + "?" + rawQuery
	}

	if err := kcontext.GetRequestError(request.Context()); err != nil {
		info.Error = err.Error()
	}

	info.StatusCode = status
	info.Time = now.Format(time.RFC3339)
	info.Cost = cost.Seconds()
	info.Duration = cost.String()
	info.ClientIP = clientIp
	info.Method = request.Method
	info.RequestID = itrace.FromContext(request.Context())
	info.RemoteHost = header.Get("Host")
	if len(info.RemoteHost) == 0 {
		info.RemoteHost = request.Host
	}

	info.RefererService = request.Header.Get(kcontext.HeaderRefererService)
	info.Referer = request.Referer()
	info.Agent = request.UserAgent()

	info.Version = request.Proto

	// ipv4:port 只保留ip地址 TODO:兼容ipv6
	if info.RemoteAddr, _, _ = net.SplitHostPort(request.RemoteAddr); info.RemoteAddr == info.ClientIP {
		info.RemoteAddr = ""
	}
	info.Uri = request.URL.Path
	info.Bytes = size
	info.Request = reqPath
	return info
}
