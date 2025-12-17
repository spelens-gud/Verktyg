package redisx

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

type redisLogger struct{}

func (redisLogger) Printf(ctx context.Context, msg string, v ...interface{}) {
	logger.FromContext(ctx).AddCallerSkip(1).Infof(msg, v...)
}

func init() {
	redis.SetLogger(redisLogger{})
}
