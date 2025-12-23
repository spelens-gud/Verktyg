package otrace

import (
	"context"
	"fmt"
	"io"
	"net/textproto"

	"github.com/opentracing/opentracing-go"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

var (
	_ itrace.Tracer = &Tracer{}
	_ itrace.Span   = &Span{}
)

type Tracer struct {
	tracer opentracing.Tracer
	closer io.Closer
}

func NewOpenTracingTracer(tracer opentracing.Tracer, closer io.Closer, err error) (itrace.Tracer, error) {
	if err != nil {
		return nil, err
	}
	return &Tracer{tracer: tracer, closer: closer}, nil
}

func (t Tracer) Close() error {
	return t.closer.Close()
}

func (t Tracer) String() string {
	return fmt.Sprintf("opentracing : %T", t.tracer)
}

func (t Tracer) Init() {
	if t.tracer != nil {
		opentracing.SetGlobalTracer(t.tracer)
	} else {
		panic("nil opentracing tracer")
	}
}

func (t Tracer) StartExitSpan(ctx context.Context, name, peer string) (span itrace.Span) {
	return t.startExitSpan(ctx, name, peer)
}

func (t Tracer) startExitSpan(ctx context.Context, name, peer string) (span *Span) {
	oSpan, _ := t.startSpanFromContext(ctx, name)
	span = &Span{Span: oSpan}
	return
}

func (t Tracer) StartSpan(ctx context.Context, name string) (span itrace.Span, nCtx context.Context) {
	oSpan, nCtx := t.startSpanFromContext(ctx, name)
	span = &Span{Span: oSpan}
	return
}

func (t Tracer) startSpanFromContext(ctx context.Context, name string, opts ...opentracing.StartSpanOption) (span opentracing.Span, nCtx context.Context) {
	span, nCtx = opentracing.StartSpanFromContext(ctx, name, opts...)

	if traceID := itrace.FromContext(ctx); len(traceID) > 0 && getTraceIDFromSpanContext(span.Context()) != traceID {
		updateSpanTraceID(span, traceID)
	}
	return
}

func (t Tracer) InjectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span itrace.Span) {
	return t.injectMetadata(ctx, name, peer, metadata, extReqIDKeys...)
}

func (t Tracer) injectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span *Span) {
	itrace.SetMetadataRequestID(itrace.FromContext(ctx), metadata, extReqIDKeys...)
	sp := t.startExitSpan(ctx, name, peer)
	// nolint
	_ = opentracing.GlobalTracer().Inject(sp.Span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(metadata))
	return sp
}

func (t Tracer) ExtractFromMetadata(ctx context.Context, name string, metadata map[string][]string, extReqIDKeys ...string) (nSpan itrace.Span, nCtx context.Context) {
	return t.extractFromMetadata(ctx, name, metadata, nil, extReqIDKeys...)
}

func (t Tracer) extractFromMetadata(
	ctx context.Context,
	name string,
	metadata map[string][]string,
	opts []opentracing.StartSpanOption,
	extReqIDKeys ...string,
) (sp *Span, nCtx context.Context) {
	var (
		oTracer    = opentracing.GlobalTracer()
		header     = textproto.MIMEHeader(metadata)
		xRequestID = itrace.GetRequestIDFromMetadata(metadata, extReqIDKeys...)
	)

	// 检查有无上游调用
	if upstreamCtx, err := oTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header)); err == nil {
		opts = append(opts, opentracing.ChildOf(upstreamCtx))
	}

	// 创建span
	oSpan, ctx := t.startSpanFromContext(ctx, name, opts...)

	if len(xRequestID) > 0 {
		oSpan.SetTag(itrace.HeaderXRequestID, xRequestID)
	}

	// 如果 xRequestID 为空 或 无法设置span的traceID时 使用新的span自动生成的traceID
	xRequestID = updateSpanTraceID(oSpan, xRequestID)

	return &Span{Span: oSpan}, itrace.WithContext(ctx, xRequestID)
}
