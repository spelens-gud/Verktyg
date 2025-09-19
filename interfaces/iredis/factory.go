package iredis

import (
	"git.bestfulfill.tech/devops/go-core/kits/kstruct"
)

type (
	// deprecated: 废弃 不建议使用
	MultiRedisClusterFactory func(cfg MultiRedisClusterConfig, multiDB interface{}) (cf func(), err error)
	// deprecated: 废弃 不建议使用
	RedisClusterConstructor func(config RedisClusterConfig) (Redis, error)

	// deprecated: 废弃 不建议使用
	MultiRedisClientFactory func(cfg MultiRedisConfig, multiDB interface{}) (cf func(), err error)
	// deprecated: 废弃 不建议使用
	RedisClientConstructor func(config RedisConfig) (Redis, error)
)

// deprecated: 废弃 不建议使用
func NewMultiRedisClusterFactory(f RedisClusterConstructor) MultiRedisClusterFactory {
	return func(cfg MultiRedisClusterConfig, multiDB interface{}) (cf func(), err error) {
		cf, err = kstruct.MultiFieldsBuilder(cfg, multiDB, "redis", func(i interface{}) (ret interface{}, err error) {
			return f(i.(RedisClusterConfig))
		}, func(i interface{}) func() {
			return func() {
				_ = i.(Redis).Close()
			}
		})
		return
	}
}

// deprecated: 废弃 不建议使用
func NewMultiRedisClientFactory(f RedisClientConstructor) MultiRedisClientFactory {
	return func(cfg MultiRedisConfig, multiDB interface{}) (cf func(), err error) {
		cf, err = kstruct.MultiFieldsBuilder(cfg, multiDB, "redis", func(i interface{}) (ret interface{}, err error) {
			return f(i.(RedisConfig))
		}, func(i interface{}) func() {
			return func() {
				_ = i.(Redis).Close()
			}
		})
		return
	}
}
