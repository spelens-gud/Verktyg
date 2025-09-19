package logger

import (
	"context"
	"errors"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger/internal"
	"git.bestfulfill.tech/devops/go-core/kits/ktrace/tracer"
)

type tracePatch struct {
	ctx context.Context
	ilog.Logger
}

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
