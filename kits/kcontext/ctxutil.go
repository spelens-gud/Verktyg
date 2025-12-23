package kcontext

import (
	"context"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

type contentKey int

const (
	metadataKey contentKey = iota
	serviceCodeKey
	requestErrorKey
)

// Deprecated: use Detach instead
//func InheritTraceSpan(ctx context.Context) context.Context {
//	return Detach(ctx)
//}

func WithNewRequestID(ctx context.Context) context.Context {
	return itrace.WithContext(ctx, itrace.NewIDFunc())
}

func GetMetadata(ctx context.Context) (data Metadata) {
	data, _ = ctx.Value(metadataKey).(Metadata)
	return
}

func SetMetadata(ctx context.Context, data Metadata) context.Context {
	return context.WithValue(ctx, metadataKey, data)
}

func GetRequestID(ctx context.Context) string {
	return itrace.FromContext(ctx)
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return itrace.WithContext(ctx, id)
}

func GetRequestError(ctx context.Context) (err error) {
	err, _ = ctx.Value(requestErrorKey).(error)
	return
}

func SetServiceCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, serviceCodeKey, code)
}

func GetServiceCode(ctx context.Context) (code int) {
	code, _ = ctx.Value(serviceCodeKey).(int)
	return
}
