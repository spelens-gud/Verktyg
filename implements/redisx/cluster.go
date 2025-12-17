package redisx

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"github.com/spelens-gud/Verktyg/interfaces/iredis"
	"github.com/spelens-gud/Verktyg/kits/kdb"
)

var _ iredis.Redis = &redisClusterClient{}

type redisClusterClient struct {
	*redis.ClusterClient
}

func (r *redisClusterClient) Raw() redis.UniversalClient {
	return r.ClusterClient
}

func (r *redisClusterClient) GetRedis(ctx context.Context) redis.Cmdable {
	// nolint
	return r.ClusterClient.WithContext(ctx)
}

func (r *redisClusterClient) Close() error {
	return r.ClusterClient.Close()
}

func NewRedisCluster(config *iredis.RedisClusterConfig, opts ...func(*redis.ClusterOptions)) (iredis.Redis, error) {
	cli := redis.NewClusterClient(config.Init(opts...))

	if err := kdb.PingDB(func() error { return cli.Ping(context.Background()).Err() },
		"redis集群", fmt.Sprint(config.Addrs)); err != nil {
		return nil, err
	}

	registerHookMetrics(cli, config.Addrs[0], "redis_cluster_"+config.Addrs[0], -1)

	return &redisClusterClient{ClusterClient: cli}, nil
}
