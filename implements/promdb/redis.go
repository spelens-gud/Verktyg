package promdb

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	"github.com/spelens-gud/Verktyg/interfaces/imetrics"
)

type RedisStatsGetter interface {
	PoolStats() *redis.PoolStats
}

func RegisterRedisStatsMetrics(dbName string, sg RedisStatsGetter) {
	constLabels := prometheus.Labels{"db_name": dbName}

	descGroup := imetrics.NewDescGroup(imetrics.NamespaceRedisStats, imetrics.SubsystemConnections, constLabels,
		func() []interface{} {
			return []interface{}{sg.PoolStats()}
		})

	descGroup.NewConstMetrics("hits", "The number of times free connection was found in the pool.",
		prometheus.GaugeValue,
		func(i ...interface{}) float64 {
			stats := i[0].(*redis.PoolStats)
			return float64(stats.Hits)
		})

	descGroup.NewConstMetrics("misses", "The number of times free connection was NOT found in the pool.",
		prometheus.GaugeValue,
		func(i ...interface{}) float64 {
			stats := i[0].(*redis.PoolStats)
			return float64(stats.Misses)
		})

	descGroup.NewConstMetrics("timeout", "The number of times a wait timeout occurred.",
		prometheus.GaugeValue,
		func(i ...interface{}) float64 {
			stats := i[0].(*redis.PoolStats)
			return float64(stats.Timeouts)
		})

	descGroup.NewConstMetrics("total", "The number of total connections in the pool.",
		prometheus.GaugeValue,
		func(i ...interface{}) float64 {
			stats := i[0].(*redis.PoolStats)
			return float64(stats.TotalConns)
		})

	descGroup.NewConstMetrics("idle", "The number of idle connections in the pool.",
		prometheus.GaugeValue,
		func(i ...interface{}) float64 {
			stats := i[0].(*redis.PoolStats)
			return float64(stats.IdleConns)
		})

	descGroup.NewConstMetrics("stale", "The number of stale connections removed from the pool.",
		prometheus.GaugeValue,
		func(i ...interface{}) float64 {
			stats := i[0].(*redis.PoolStats)
			return float64(stats.StaleConns)
		})

	imetrics.MustRegister(descGroup)
}
