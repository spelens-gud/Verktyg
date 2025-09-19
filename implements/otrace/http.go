package otrace

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"git.bestfulfill.tech/devops/go-core/interfaces/itrace"
)

func (t Tracer) InjectHttp(req *http.Request, extReqIDKeys ...string) (span itrace.Span) {
	var (
		ctx  = req.Context()
		name = itrace.InjectHttpName(req)
		peer = itrace.InjectHttpPeer(req)
	)

	span = t.InjectMetadata(ctx, name, peer, req.Header)
	itrace.SetHttpClientTag(span, req)
	return
}

func (t Tracer) ExtractHttp(req *http.Request, extReqIDKeys ...string) (span itrace.Span, nCtx context.Context) {
	return t.extractHttpRequest(req, extReqIDKeys...)
}

func (t Tracer) extractHttpRequest(req *http.Request, extReqIDKeys ...string) (span *Span, ctx context.Context) {
	opts := []opentracing.StartSpanOption{opentracing.Tag{Key: string(ext.Component), Value: "HTTP"}}
	span, ctx = t.extractFromMetadata(req.Context(), itrace.ExtractHttpName(req), req.Header, opts, extReqIDKeys...)
	itrace.SetHttpServerTag(span, req)
	return
}
