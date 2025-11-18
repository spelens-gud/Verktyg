package promdb

import (
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

// BenchmarkRegisterRedisStatsMetrics 基准测试注册 Redis 统计指标
func BenchmarkRegisterRedisStatsMetrics(b *testing.B) {
	stats := &redis.PoolStats{
		Hits:       1000,
		Misses:     100,
		Timeouts:   10,
		TotalConns: 50,
		IdleConns:  40,
		StaleConns: 5,
	}

	mock := &mockRedisStatsGetter{stats: stats}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 使用不同的数据库名称避免重复注册
		RegisterRedisStatsMetrics(fmt.Sprintf("bench_db_%d", i), mock)
	}
}

// BenchmarkPoolStats 基准测试获取连接池统计信息
func BenchmarkPoolStats(b *testing.B) {
	stats := &redis.PoolStats{
		Hits:       1000,
		Misses:     100,
		Timeouts:   10,
		TotalConns: 50,
		IdleConns:  40,
		StaleConns: 5,
	}

	mock := &mockRedisStatsGetter{stats: stats}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mock.PoolStats()
	}
}

// BenchmarkRegisterRedisStatsMetrics_SmallPool 基准测试小连接池
func BenchmarkRegisterRedisStatsMetrics_SmallPool(b *testing.B) {
	stats := &redis.PoolStats{
		Hits:       10,
		Misses:     1,
		Timeouts:   0,
		TotalConns: 5,
		IdleConns:  4,
		StaleConns: 0,
	}

	mock := &mockRedisStatsGetter{stats: stats}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 使用不同的数据库名称避免重复注册
		RegisterRedisStatsMetrics(fmt.Sprintf("small_pool_%d", i), mock)
	}
}

// BenchmarkRegisterRedisStatsMetrics_LargePool 基准测试大连接池
func BenchmarkRegisterRedisStatsMetrics_LargePool(b *testing.B) {
	stats := &redis.PoolStats{
		Hits:       1000000,
		Misses:     50000,
		Timeouts:   100,
		TotalConns: 500,
		IdleConns:  450,
		StaleConns: 10,
	}

	mock := &mockRedisStatsGetter{stats: stats}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 使用不同的数据库名称避免重复注册
		RegisterRedisStatsMetrics(fmt.Sprintf("large_pool_%d", i), mock)
	}
}

// BenchmarkConcurrentPoolStats 基准测试并发获取连接池统计
func BenchmarkConcurrentPoolStats(b *testing.B) {
	stats := &redis.PoolStats{
		Hits:       1000,
		Misses:     100,
		Timeouts:   10,
		TotalConns: 50,
		IdleConns:  40,
		StaleConns: 5,
	}

	mock := &mockRedisStatsGetter{stats: stats}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mock.PoolStats()
		}
	})
}
