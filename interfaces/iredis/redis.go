package iredis

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"net"
	"runtime"
	"time"

	"git.bestfulfill.tech/devops/go-core/kits/knet"
)

type (
	Redis interface {
		Raw() redis.UniversalClient
		GetRedis() redis.Cmdable
		Close() error
	}

	// deprecated: 废弃 不建议使用
	MultiRedisConfig map[string]RedisConfig
	// deprecated: 废弃 不建议使用
	MultiRedisClusterConfig map[string]RedisClusterConfig

	BaseConfig struct {
		// Password 密码
		Password string `json:"password"`
		// PoolSize 集群内每个节点的最大连接池限制 默认每个CPU10个连接
		PoolSize int `json:"pool_size"`
		// MaxRetries 网络相关的错误最大重试次数 默认2次
		MaxRetries int `json:"max_retries"`
		// MinIdleConns 最小空闲连接数
		MinIdleConns int `json:"min_idle_conns"`
		// DialTimeout 拨超时时间 单位：毫秒
		DialTimeout int `json:"dial_timeout"`
		// ReadTimeout 读超时 默认3s 单位：毫秒
		ReadTimeout int `json:"read_timeout"`
		// WriteTimeout 读超时 默认3s 单位：毫秒
		WriteTimeout int `json:"write_timeout"`
		// IdleTimeout 连接最大空闲时间，默认60s, 超过该时间，连接会被主动关闭 单位：毫秒
		IdleTimeout int `json:"idle_timeout"`
		// 是否开启链路追踪
		EnableTrace bool `json:"enable_trace"`
		// 慢日志阀值，超过该阀值的请求，将被记录到慢日志中
		SlowThreshold int `json:"slow_threshold"`
		// tls配置
		InsecureSkipVerify bool `json:"insecure_skip_verify"`
	}

	RedisConfig struct {
		BaseConfig
		// Addr 节点连接地址
		Addr string `json:"addr"`
		// DB
		DB int `json:"db"`
	}

	RedisClusterConfig struct {
		BaseConfig
		// Addrs 集群实例配置地址
		Addrs []string `json:"addrs"`
		// MaxRedirects 集群重定向最大次数
		MaxRedirects int `json:"max_redirects"`
	}
)

func (cfg *BaseConfig) Init() {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 20 + runtime.GOMAXPROCS(-1)*20
	}

	if cfg.MinIdleConns <= 0 {
		cfg.MinIdleConns = 20
	}

	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 2
	}

	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 500
	}

	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 3 * 1000
	}

	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 3 * 1000
	}

	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 60 * 1000
	}
}

func (config *RedisClusterConfig) Init(opts ...func(options *redis.ClusterOptions)) *redis.ClusterOptions {
	config.BaseConfig.Init()

	if config.MaxRedirects == 0 {
		config.MaxRedirects = 8
	}

	clusterOpt := &redis.ClusterOptions{
		Addrs:        config.Addrs,
		MaxRedirects: config.MaxRedirects,
		Password:     config.Password,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  time.Millisecond * time.Duration(config.DialTimeout),
		ReadTimeout:  time.Millisecond * time.Duration(config.ReadTimeout),
		WriteTimeout: time.Millisecond * time.Duration(config.WriteTimeout),
		PoolTimeout:  time.Millisecond * time.Duration(config.DialTimeout),
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		IdleTimeout:  time.Millisecond * time.Duration(config.IdleTimeout),
	}

	if config.InsecureSkipVerify {
		clusterOpt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	clusterOpt.Dialer = redisDialer(clusterOpt.DialTimeout, clusterOpt.TLSConfig)

	for _, o := range opts {
		o(clusterOpt)
	}

	return clusterOpt
}

func (config *RedisConfig) Init(opts ...func(options *redis.Options)) *redis.Options {
	config.BaseConfig.Init()

	redisOpt := &redis.Options{
		DB:           config.DB,
		Addr:         config.Addr,
		Password:     config.Password,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  time.Millisecond * time.Duration(config.DialTimeout),
		ReadTimeout:  time.Millisecond * time.Duration(config.ReadTimeout),
		WriteTimeout: time.Millisecond * time.Duration(config.WriteTimeout),
		PoolTimeout:  time.Millisecond * time.Duration(config.DialTimeout),
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		IdleTimeout:  time.Millisecond * time.Duration(config.IdleTimeout),
	}

	if config.InsecureSkipVerify {
		redisOpt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	redisOpt.Dialer = redisDialer(redisOpt.DialTimeout, redisOpt.TLSConfig)

	for _, o := range opts {
		o(redisOpt)
	}
	return redisOpt
}

func redisDialer(timeout time.Duration, tlsConfig *tls.Config) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return knet.WrapTcpDialer(timeout, 5*time.Minute, tlsConfig)
}
