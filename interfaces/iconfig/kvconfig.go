package iconfig

import (
	"context"
	"time"
)

type (
	KVConfigLoader interface {
		Get(key string) (value string, err error)

		Set(key string, value string, exp time.Duration) (err error)

		GetWithContext(ctx context.Context, key string) (value string, err error)

		SetWithContext(ctx context.Context, key string, value string, exp time.Duration) (err error)
	}

	KVConfigSourceLoader interface {
		GetWithContext(ctx context.Context, key string) (value string, err error)

		SetWithContext(ctx context.Context, key string, value string) (err error)
	}
)
