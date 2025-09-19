package kdb

import (
	"time"

	"git.bestfulfill.tech/devops/go-core/implements/promdb"
	"git.bestfulfill.tech/devops/go-core/implements/prommq"
	"git.bestfulfill.tech/devops/go-core/interfaces/imetrics"
)

var (
	DefaultSqlMetrics       = promdb.InitDBUsageMetrics("mysql", imetrics.NamespaceSql)
	DefaultRedisMetrics     = promdb.InitDBUsageMetrics("redis", imetrics.NamespaceRedis)
	DefaultKafkaMetrics     = prommq.InitMQUsageMetrics("kafka")
	DefaultAmqpMetrics      = prommq.InitMQUsageMetrics("amqp")
	DefaultRedisListMetrics = prommq.InitMQUsageMetrics("redis_list")
)

func SqlMetrics(name, addr, command string, err error, duration time.Duration) {
	DefaultSqlMetrics.Metrics(name, addr, command, err, duration)
}

func RedisMetrics(name, addr, command string, err error, duration time.Duration) {
	DefaultRedisMetrics.Metrics(name, addr, command, err, duration)
}

func KafkaProduceMetrics(addr, topic string, err error, duration time.Duration) {
	DefaultKafkaMetrics.MetricsProduce(addr, topic, err, duration)
}

func KafkaConsumeMetrics(addr, topic string, err error, duration time.Duration) {
	DefaultKafkaMetrics.MetricsConsume(addr, topic, err, duration)
}

func AmqpProduceMetrics(addr, topic string, err error, duration time.Duration) {
	DefaultAmqpMetrics.MetricsProduce(addr, topic, err, duration)
}

func AmqpConsumeMetrics(addr, topic string, err error, duration time.Duration) {
	DefaultAmqpMetrics.MetricsConsume(addr, topic, err, duration)
}

func RedisListProduceMetrics(addr, topic string, err error, duration time.Duration) {
	DefaultRedisListMetrics.MetricsProduce(addr, topic, err, duration)
}

func RedisListConsumeMetrics(addr, topic string, err error, duration time.Duration) {
	DefaultRedisListMetrics.MetricsConsume(addr, topic, err, duration)
}
