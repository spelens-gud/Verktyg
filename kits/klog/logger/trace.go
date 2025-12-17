package logger

import (
	"context"
	"errors"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger/internal"
	"github.com/spelens-gud/Verktyg.git/kits/ktrace/tracer"
)

type tracePatch struct {
	ctx context.Context
	ilog.Logger
}

// nolint
func (_ tracePatch) Patch(ctx context.Context, logger ilog.Logger) ilog.Logger {
	if ctx == ctxBackground || tracer.IsNoop() {
		return logger
	}
	tl := &tracePatch{
		ctx: ctx,
	}
	tl.Logger = internal.WrapLogger(logger, tl.traceLog)
	return tl
}

// 链路追踪日志
func (t *tracePatch) traceLog(level ilog.Level, msg string, args ...interface{}) bool {
	msg = internal.Sprintf(msg, args...)
	if level >= ilog.Error {
		tracer.SpanFromContext(t.ctx).Error(errors.New(msg), t.FieldData())
	} else {
		tracer.SpanFromContext(t.ctx).Log(level.String(), msg, t.FieldData())
	}
	return true
}
