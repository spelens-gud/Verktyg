package redisx

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"

	"git.bestfulfill.tech/devops/go-core/interfaces/iredis"
	"git.bestfulfill.tech/devops/go-core/kits/kdb"
)

var _ iredis.Redis = &redisClient{}

type redisClient struct {
	*redis.Client
}

func (r *redisClient) Raw() redis.UniversalClient {
	return r.Client
}

func (r *redisClient) GetRedis() redis.Cmdable {
	// nolint
	return r.Client
}

func (r *redisClient) Close() error {
	return r.Client.Close()
}

func NewRedis(config *iredis.RedisConfig, opts ...func(*redis.Options)) (iredis.Redis, error) {
	cli := redis.NewClient(config.Init(opts...))

	if err := kdb.PingDB(func() error { return cli.Ping(context.Background()).Err() },
		"redis", fmt.Sprint(config.Addr, ":", config.DB)); err != nil {
		return nil, err
	}

	registerHookMetrics(cli, config.Addr, config.Addr+"_"+strconv.Itoa(config.DB), config.DB)

	return &redisClient{Client: cli}, nil
}
