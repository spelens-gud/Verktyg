package redisx

import (
	"context"

	"github.com/redis/go-redis/v9"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type redisLogger struct{}

func (redisLogger) Printf(ctx context.Context, msg string, v ...interface{}) {
	logger.FromContext(ctx).AddCallerSkip(1).Infof(msg, v...)
}

func init() {
	redis.SetLogger(redisLogger{})
}
