package skytrace

import (
	"context"
	"errors"
	"net/textproto"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter"

	"github.com/spelens-gud/Verktyg/implements/skytrace/samplereport"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/itrace"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

var (
	_              itrace.Tracer = &Tracer{}
	_              itrace.Span   = &Span{}
	noopSpan                     = &go2sky.NoopSpan{}
	traceLogOption               = reporter.WithLogger(logger.NewStandardLogger())
)

type Tracer struct {
	*go2sky.Tracer
	reporter go2sky.Reporter
}

func (t Tracer) String() string {
	return "go2sky"
}

type TraceOption struct {
	ReportOptions []reporter.GRPCReporterOption
	SampleRate    float64
}

type Option func(opt *TraceOption)

func NewLogTracer(name string, lg Logger) (tracer itrace.Tracer) {
	if len(name) == 0 {
		name = "service"
	}
	rep := NewLogReporter(lg)
	skyTracer, _ := go2sky.NewTracer(name, go2sky.WithReporter(rep))
	return &Tracer{Tracer: skyTracer, reporter: rep}
}

func (t Tracer) Close() error {
	t.reporter.Close()
	return nil
}

func NewGrpcTracer(name string, reportAddr string, opts ...Option) (tracer itrace.Tracer, err error) {
	if len(name) == 0 {
		name = iconfig.GetApplicationName()
	}

	if len(reportAddr) == 0 {
		return itrace.NoopTracer, nil
	}

	opt := &TraceOption{
		ReportOptions: []reporter.GRPCReporterOption{traceLogOption},
	}
	for _, o := range opts {
		o(opt)
	}

	rep, err := reporter.NewGRPCReporter(reportAddr, opt.ReportOptions...)
	if err != nil {
		return
	}

	tracerOpt := []go2sky.TracerOption{go2sky.WithInstance(iconfig.HostName())}

	// go2sky采样率不可用 使用上报侧过滤
	if opt.SampleRate > 0 {
		if opt.SampleRate > 1 {
			err = errors.New("invalid sample rate")
			return
		}
		tracerOpt = append(tracerOpt, go2sky.WithReporter(samplereport.NewReporter(rep, opt.SampleRate)))
	} else {
		tracerOpt = append(tracerOpt, go2sky.WithReporter(rep))
	}

	skyTracer, err := go2sky.NewTracer(name, tracerOpt...)
	if err != nil {
		return nil, err
	}
	return &Tracer{Tracer: skyTracer, reporter: rep}, nil
}

func (t Tracer) Init() {}

func (t Tracer) StartSpan(ctx context.Context, name string) (span itrace.Span, nCtx context.Context) {
	sp, nCtx := t.startSpan(ctx, name)

	if traceID := itrace.FromContext(ctx); len(traceID) > 0 && go2sky.TraceID(nCtx) != traceID {
		setSpanTraceID(sp.Span, traceID)
	}

	return sp, nCtx
}

func (t Tracer) startSpan(ctx context.Context, name string, opts ...go2sky.SpanOption) (span *Span, nCtx context.Context) {
	sp, nCtx, err := t.CreateLocalSpan(ctx, opts...)
	if err != nil {
		sp = noopSpan
	}
	if nCtx == nil {
		nCtx = ctx
	}
	sp.SetOperationName(name)
	span = &Span{Span: sp}
	return
}

func (t Tracer) ExtractFromMetadata(ctx context.Context, name string, metadata map[string][]string, extReqIDKeys ...string) (nSpan itrace.Span, nCtx context.Context) {
	return t.extractFromMetadata(ctx, name, metadata, extReqIDKeys...)
}

func (t Tracer) extractFromMetadata(ctx context.Context, name string, metadata map[string][]string, extReqIDKeys ...string) (nSpan *Span, nCtx context.Context) {
	// 获取头部自定义请求ID
	xRequestID := itrace.GetRequestIDFromMetadata(metadata, extReqIDKeys...)

	// 从sw8提取上游
	sp, nCtx, err := t.CreateEntrySpan(ctx, name, func() (s string, err error) {
		return textproto.MIMEHeader(metadata).Get(propagation.Header), nil
	})

	// 返回err说明头部获取不到 上游的的调用 且ctx会为空
	if err != nil {
		nSpan, nCtx = t.startSpan(ctx, name, go2sky.WithSpanType(go2sky.SpanTypeEntry))
	} else {
		nSpan = &Span{Span: sp}
	}

	// 如果有自定义的请求ID 则使用自定义ID 否则将xRequestID置为新建span的ID
	if len(xRequestID) > 0 && go2sky.TraceID(nCtx) != xRequestID {
		nSpan.Tag(itrace.HeaderXRequestID, xRequestID)
		setSpanTraceID(nSpan.Span, xRequestID)
	} else {
		xRequestID = go2sky.TraceID(nCtx)
	}

	return nSpan, itrace.WithContext(nCtx, xRequestID)
}

func (t Tracer) InjectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span itrace.Span) {
	return t.injectMetadata(ctx, name, peer, metadata, extReqIDKeys...)
}

func (t Tracer) StartExitSpan(ctx context.Context, name, peer string) (span itrace.Span) {
	return t.startExitSpan(ctx, name, peer, noopInject)
}

func noopInject(header string) error {
	return nil
}

func (t Tracer) startExitSpan(ctx context.Context, name, peer string, inject func(header string) error) (span *Span) {
	traceID := itrace.FromContext(ctx)

	if inject == nil {
		inject = noopInject
	}

	sp, err := t.CreateExitSpan(ctx, name, peer, inject)

	if err != nil {
		sp = noopSpan
	} else if len(traceID) > 0 && getSpanTraceID(sp) != traceID {
		setSpanTraceID(sp, traceID)
	}

	span = &Span{Span: sp}
	return
}

func (t Tracer) injectMetadata(ctx context.Context, name, peer string, metadata map[string][]string, extReqIDKeys ...string) (span *Span) {
	traceID := itrace.FromContext(ctx)
	return t.startExitSpan(ctx, name, peer, func(header string) error {
		textproto.MIMEHeader(metadata).Add(propagation.Header, header)
		itrace.SetMetadataRequestID(traceID, metadata, extReqIDKeys...)
		return nil
	})
}
