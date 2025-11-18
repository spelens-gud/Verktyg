package redistorew

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"git.bestfulfill.tech/devops/go-core/interfaces/iredis"
	"git.bestfulfill.tech/devops/go-core/kits/kgo/cachew"
)

type redisW struct {
	Redis iredis.Redis
}

func (r redisW) Delete(ctx context.Context, key string) (err error) {
	return r.Redis.GetRedis().Del(ctx, key).Err()
}

func (r redisW) Get(ctx context.Context, key string) ([]byte, error) {
	return r.Redis.GetRedis().Get(ctx, key).Bytes()
}

func (r redisW) Set(ctx context.Context, key string, data []byte, exp time.Duration) error {
	return r.Redis.GetRedis().Set(ctx, key, data, exp).Err()
}

func (r redisW) KeyNotFound(err error) bool {
	return err == redis.Nil
}

func NewIRedisStoreW(redis iredis.Redis) cachew.Store {
	return redisW{Redis: redis}
}
