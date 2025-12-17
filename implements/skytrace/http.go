package skytrace

import (
	"context"
	"net/http"

	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"

	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
)

func (t Tracer) ExtractHttp(req *http.Request, extReqIDKeys ...string) (span itrace.Span, nCtx context.Context) {
	return t.extractHttp(req, extReqIDKeys...)
}

func (t Tracer) extractHttp(req *http.Request, extReqIDKeys ...string) (nSpan *Span, nCtx context.Context) {
	var (
		name = itrace.ExtractHttpName(req)
		ctx  = req.Context()
	)

	nSpan, nCtx = t.extractFromMetadata(ctx, name, req.Header, extReqIDKeys...)
	nSpan.SetSpanLayer(v3.SpanLayer_Http)
	itrace.SetHttpServerTag(nSpan, req)
	return
}

func (t Tracer) InjectHttp(req *http.Request, extReqIDKeys ...string) (span itrace.Span) {
	return t.injectHttp(req, extReqIDKeys...)
}

func (t Tracer) injectHttp(req *http.Request, extReqIDKeys ...string) (span *Span) {
	var (
		ctx  = req.Context()
		peer = itrace.InjectHttpPeer(req)
		name = itrace.InjectHttpName(req)
	)

	span = t.injectMetadata(ctx, name, peer, req.Header, extReqIDKeys...)
	span.SetLayer(itrace.SpanLayerHttp)
	span.SetPeer(peer)
	itrace.SetHttpClientTag(span, req)
	return
}
