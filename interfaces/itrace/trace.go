package itrace

import (
	"context"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/google/uuid"

	"github.com/spelens-gud/Verktyg/internal/incontext"
)

func NewIDFunc() string {
	return newIDFunc()
}

func SetIDFunc(f func() string) {
	newIDFunc = f
}

var newIDFunc = func() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

const keyTraceID = incontext.Key("context.trace_id")

type (
	TracerConfig interface {
		NewTracer() (tracer Tracer, err error)

		Names() []string
	}

	Tracer interface {
		// 返回名
		String() string

		// 初始化
		Init()

		// 关闭
		Close() error

		// 发起新分片
		StartSpan(ctx context.Context, name string) (span Span, nCtx context.Context)

		// 发起新分片
		StartExitSpan(ctx context.Context, name, peer string) (span Span)

		// 发起http下游调用
		InjectHttp(req *http.Request, extReqIDKeys ...string) (span Span)

		// 从上游http导出
		ExtractHttp(req *http.Request, extReqIDKeys ...string) (span Span, nCtx context.Context)

		// 从metadata导出
		ExtractFromMetadata(ctx context.Context, name string, metadata map[string][]string, extReqIDKeys ...string) (nSpan Span, nCtx context.Context)

		// 导入到metadata
		InjectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span Span)
	}

	Span interface {
		SetOperationName(string)

		SetComponent(int)

		SetPeer(string)

		SetLayer(int)

		Tag(key, value string)

		Finish()

		Error(err error, fields map[string]interface{})

		Log(level, msg string, fields map[string]interface{})
	}
)

type LogField struct {
	Key   string
	Value interface{}
}

const (
	SpanLayerUnknown      = 0
	SpanLayerDatabase     = 1
	SpanLayerRPCFramework = 2
	SpanLayerHttp         = 3
	SpanLayerMQ           = 4
	SpanLayerCache        = 5
)

const (
	HeaderXRequestID = "X-Request-ID"
	HeaderXTraceID   = "X-Trace-ID"
)

var DefaultReqIDKeys = []string{HeaderXRequestID, HeaderXTraceID}

func SetMetadataRequestID(requestID string, metadata map[string][]string, extReqIDKeys ...string) {
	if len(requestID) == 0 {
		return
	}
	if len(extReqIDKeys) == 0 {
		extReqIDKeys = DefaultReqIDKeys
	}
	for _, key := range extReqIDKeys {
		textproto.MIMEHeader(metadata).Set(key, requestID)
	}
}

func GetRequestIDFromMetadata(metadata map[string][]string, extReqIDKeys ...string) string {
	if len(extReqIDKeys) == 0 {
		extReqIDKeys = DefaultReqIDKeys
	}
	var requestID string
	for _, key := range extReqIDKeys {
		if requestID = textproto.MIMEHeader(metadata).Get(key); len(requestID) > 0 {
			break
		}
	}
	return requestID
}

var spanLayerName = map[int]string{
	SpanLayerUnknown:      "Unknown",
	SpanLayerDatabase:     "Database",
	SpanLayerRPCFramework: "RPCFramework",
	SpanLayerHttp:         "Http",
	SpanLayerMQ:           "MQ",
	SpanLayerCache:        "Cache",
}

func GetSpanLayerName(i int) string {
	return spanLayerName[i]
}

func FromContext(ctx context.Context) string {
	id, ok := keyTraceID.Value(ctx).(string)
	if !ok {
		return ""
	}
	return id
}

func WithContext(ctx context.Context, id string) (nCtx context.Context) {
	if len(id) == 0 {
		id = NewIDFunc()
	}
	nCtx = keyTraceID.WithValue(ctx, id)
	return
}

func ExtractHttpName(req *http.Request) (name string) {
	return "HTTP_SERVER:" + req.Method + ":" + req.Host
}

func InjectHttpName(req *http.Request) (name string) {
	return "HTTP_CLIENT:" + req.Method + ":" + req.URL.Host
}

func InjectHttpPeer(req *http.Request) (peer string) {
	return req.URL.Host
}
