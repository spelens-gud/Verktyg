package logger

import (
	"context"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
)

type FieldPatch map[string]func(ctx context.Context) interface{}

// 传入contextKey 自动提取
func NewContextFieldPatch(m map[string]interface{}) ilog.LoggerPatch {
	p := make(FieldPatch, len(m))
	for k, v := range m {
		v := v
		p[k] = func(ctx context.Context) interface{} {
			if ctx == ctxBackground {
				return nil
			}
			return ctx.Value(v)
		}
	}
	return p
}

func (f FieldPatch) Patch(ctx context.Context, logger ilog.Logger) ilog.Logger {
	var (
		fields map[string]interface{}
		ignore = 0
		keyLen = len(f)
	)

	for k, v := range f {
		value := v(ctx)
		if value == nil {
			ignore += 1
			continue
		}
		if fields == nil {
			fields = make(map[string]interface{}, keyLen-ignore)
		}
		fields[k] = value
	}

	if len(fields) == 0 {
		return logger
	}

	return logger.WithFields(fields)
}
