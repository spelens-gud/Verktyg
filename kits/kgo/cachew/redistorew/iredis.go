package redistorew

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/spelens-gud/Verktyg.git/interfaces/iredis"
	"github.com/spelens-gud/Verktyg.git/kits/kgo/cachew"
)

type redisW struct {
	Redis iredis.Redis
}

func (r redisW) Delete(ctx context.Context, key string) (err error) {
	return r.Redis.GetRedis(ctx).Del(ctx, key).Err()
}

func (r redisW) Get(ctx context.Context, key string) ([]byte, error) {
	return r.Redis.GetRedis(ctx).Get(ctx, key).Bytes()
}

func (r redisW) Set(ctx context.Context, key string, data []byte, exp time.Duration) error {
	return r.Redis.GetRedis(ctx).Set(ctx, key, data, exp).Err()
}

func (r redisW) KeyNotFound(err error) bool {
	return err == redis.Nil
}

func NewIRedisStoreW(redis iredis.Redis) cachew.Store {
	return redisW{Redis: redis}
}
