package otrace

import (
	"reflect"
	"unsafe"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

// 设置自定义traceID到context 成功返回 自定义ID 否则返回Span里的traceID
func updateSpanTraceID(span opentracing.Span, traceID string) string {
	if len(traceID) == 0 {
		return getTraceIDFromSpanContext(span.Context())
	}
	var set bool
	switch sp := span.(type) {
	case *jaeger.Span:
		id, err := jaeger.TraceIDFromString(traceID)
		if err != nil {
			break
		}
		sp.Lock()
		setJaegerSpanContextTraceID(sp, id)
		sp.Unlock()
		set = true
	}
	if set {
		return traceID
	}
	return getTraceIDFromSpanContext(span.Context())
}

func setJaegerSpanContextTraceID(sp *jaeger.Span, traceID jaeger.TraceID) {
	// nolint
	traceIDPtr := reflect.ValueOf(sp).Elem().FieldByName("context").FieldByName("traceID").UnsafeAddr()
	// nolint
	*(*jaeger.TraceID)(unsafe.Pointer(traceIDPtr)) = traceID
}

func getTraceIDFromSpanContext(ctx opentracing.SpanContext) string {
	switch c := ctx.(type) {
	case jaeger.SpanContext:
		return c.TraceID().String()
	}
	return ""
}
