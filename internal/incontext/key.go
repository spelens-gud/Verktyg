package incontext

import "context"

type (
	Key string
)

func (key Key) Value(ctx context.Context) interface{} {
	return ctx.Value(key)
}

func (key Key) WithValue(ctx context.Context, i interface{}) context.Context {
	return context.WithValue(ctx, key, i)
}
