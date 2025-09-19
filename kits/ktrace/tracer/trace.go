package tracer

import (
	"context"
	"net/http"

	"git.bestfulfill.tech/devops/go-core/interfaces/itrace"
	"git.bestfulfill.tech/devops/go-core/internal/incontext"
)

var defaultTracer = itrace.NoopTracer

const contextKeySpan incontext.Key = "tracer.span"

func IsNoop() bool { return defaultTracer == itrace.NoopTracer }

func SpanFromContext(ctx context.Context) itrace.Span {
	if ctx == nil {
		return itrace.NoopSpan
	}
	sp, _ := contextKeySpan.Value(ctx).(itrace.Span)
	if sp == nil {
		sp = itrace.NoopSpan
	}
	return sp
}

func SpanWithContext(ctx context.Context, span itrace.Span) context.Context {
	return contextKeySpan.WithValue(ctx, span)
}

func SetTracer(t itrace.Tracer) {
	if t != nil {
		t.Init()
		defaultTracer = t
	}
}

func StartSpan(ctx context.Context, name string) (span itrace.Span, nCtx context.Context) {
	span, nCtx = defaultTracer.StartSpan(ctx, name)
	nCtx = SpanWithContext(nCtx, span)
	span.Tag(itrace.TagSpanType, itrace.SpanTypeLocal)
	return
}

func StartExitSpan(ctx context.Context, name, peer string) (span itrace.Span) {
	span = defaultTracer.StartExitSpan(ctx, name, peer)
	span.SetPeer(peer)
	span.Tag(itrace.TagSpanType, itrace.SpanTypeExit)
	return
}

func InjectHttp(req *http.Request, extReqIDKeys ...string) (span itrace.Span) {
	span = defaultTracer.InjectHttp(req, extReqIDKeys...)
	span.Tag(itrace.TagSpanType, itrace.SpanTypeExit)
	return
}

func ExtractHttp(req *http.Request, extReqIDKeys ...string) (span itrace.Span, nCtx context.Context) {
	span, nCtx = defaultTracer.ExtractHttp(req, extReqIDKeys...)
	nCtx = SpanWithContext(nCtx, span)
	span.Tag(itrace.TagSpanType, itrace.SpanTypeEntry)
	return
}

func InjectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span itrace.Span) {
	span = defaultTracer.InjectMetadata(ctx, name, peer, metadata, extReqIDKeys...)
	span.Tag(itrace.TagSpanType, itrace.SpanTypeExit)
	return
}

func ExtractMetadata(ctx context.Context, name string, metadata map[string][]string, extReqIDKeys ...string) (span itrace.Span, nCtx context.Context) {
	span, nCtx = defaultTracer.ExtractFromMetadata(ctx, name, metadata, extReqIDKeys...)
	nCtx = SpanWithContext(nCtx, span)
	span.Tag(itrace.TagSpanType, itrace.SpanTypeEntry)
	return
}
