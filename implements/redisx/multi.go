package redisx

import (
	"github.com/go-redis/redis/v8"

	"github.com/spelens-gud/Verktyg/interfaces/iredis"
)

// DEPRECATED: 不要再用这种写法 请使用类型别名
func NewMultiRedisCli(cfg iredis.MultiRedisConfig, multiDB interface{}, opts ...func(*redis.Options)) (cf func(), err error) {
	return iredis.NewMultiRedisClientFactory(func(config iredis.RedisConfig) (iredis.Redis, error) {
		return NewRedis(&config, opts...)
	})(cfg, multiDB)
}

// DEPRECATED: 不要再用这种写法 请使用类型别名
func NewMultiRedisCluster(cfg iredis.MultiRedisClusterConfig, multiDB interface{}, opts ...func(*redis.ClusterOptions)) (cf func(), err error) {
	return iredis.NewMultiRedisClusterFactory(func(config iredis.RedisClusterConfig) (iredis.Redis, error) {
		return NewRedisCluster(&config, opts...)
	})(cfg, multiDB)
}
