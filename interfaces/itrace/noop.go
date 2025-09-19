package itrace

import (
	"context"
	"net/http"
)

var (
	NoopTracer Tracer = noopTracer{}
	NoopSpan   Span   = noopSpan{}
)

type (
	noopTracer struct{}
	noopSpan   struct{}
)

func (n noopSpan) Error(error, map[string]interface{})                  {}
func (n noopSpan) Log(level, msg string, fields map[string]interface{}) {}
func (n noopSpan) SetOperationName(s string)                            {}
func (n noopSpan) SetComponent(i int)                                   {}
func (n noopSpan) SetPeer(s string)                                     {}
func (n noopSpan) SetLayer(i int)                                       {}
func (n noopSpan) Tag(key, value string)                                {}
func (n noopSpan) Finish()                                              {}

func (n noopTracer) String() string {
	return "noop"
}

func (n noopTracer) Close() error { return nil }
func (n noopTracer) Init()        {}

func (n noopTracer) ExtractFromMetadata(ctx context.Context, name string, metadata map[string][]string, extReqIDKeys ...string) (nSpan Span, nCtx context.Context) {
	if traceID := GetRequestIDFromMetadata(metadata, extReqIDKeys...); len(traceID) > 0 {
		ctx = WithContext(ctx, traceID)
	} else {
		ctx = WithContext(ctx, NewIDFunc())
	}
	return NoopSpan, ctx
}

func (n noopTracer) InjectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span Span) {
	if traceID := FromContext(ctx); len(traceID) > 0 {
		SetMetadataRequestID(traceID, metadata, extReqIDKeys...)
	}
	return NoopSpan
}

func (n noopTracer) StartSpan(ctx context.Context, name string) (span Span, nCtx context.Context) {
	if len(FromContext(ctx)) == 0 {
		ctx = WithContext(ctx, NewIDFunc())
	}
	return NoopSpan, ctx
}

func (n noopTracer) StartExitSpan(ctx context.Context, name, peer string) (span Span) {
	return NoopSpan
}

func (n noopTracer) InjectHttp(req *http.Request, extReqIDKeys ...string) (span Span) {
	return n.InjectMetadata(req.Context(), "", "", req.Header, extReqIDKeys...)
}

func (n noopTracer) ExtractHttp(req *http.Request, extReqIDKeys ...string) (span Span, nCtx context.Context) {
	return n.ExtractFromMetadata(req.Context(), "", req.Header, extReqIDKeys...)
}

var _ Tracer = &noopTracer{}
