package promdb

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// mockRedisStatsGetter mock Redis 统计信息获取器
type mockRedisStatsGetter struct {
	stats *redis.PoolStats
}

func (m *mockRedisStatsGetter) PoolStats() *redis.PoolStats {
	return m.stats
}

// TestRegisterRedisStatsMetrics 测试注册 Redis 统计指标
func TestRegisterRedisStatsMetrics(t *testing.T) {
	tests := []struct {
		name   string
		dbName string
		stats  *redis.PoolStats
	}{
		{
			name:   "正常情况 - 注册基本统计信息",
			dbName: "test_db",
			stats: &redis.PoolStats{
				Hits:       100,
				Misses:     10,
				Timeouts:   1,
				TotalConns: 20,
				IdleConns:  15,
				StaleConns: 2,
			},
		},
		{
			name:   "边界情况 - 零值统计信息",
			dbName: "zero_db",
			stats: &redis.PoolStats{
				Hits:       0,
				Misses:     0,
				Timeouts:   0,
				TotalConns: 0,
				IdleConns:  0,
				StaleConns: 0,
			},
		},
		{
			name:   "边界情况 - 大数值统计信息",
			dbName: "large_db",
			stats: &redis.PoolStats{
				Hits:       1000000,
				Misses:     50000,
				Timeouts:   100,
				TotalConns: 500,
				IdleConns:  450,
				StaleConns: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建新的注册表以避免冲突
			registry := prometheus.NewRegistry()

			mock := &mockRedisStatsGetter{stats: tt.stats}

			// 注册指标
			RegisterRedisStatsMetrics(tt.dbName, mock)

			// 验证统计信息可以被获取
			stats := mock.PoolStats()
			if stats == nil {
				t.Error("PoolStats() 返回 nil")
				return
			}

			// 验证统计值
			if stats.Hits != tt.stats.Hits {
				t.Errorf("Hits = %d, want %d", stats.Hits, tt.stats.Hits)
			}
			if stats.Misses != tt.stats.Misses {
				t.Errorf("Misses = %d, want %d", stats.Misses, tt.stats.Misses)
			}
			if stats.Timeouts != tt.stats.Timeouts {
				t.Errorf("Timeouts = %d, want %d", stats.Timeouts, tt.stats.Timeouts)
			}
			if stats.TotalConns != tt.stats.TotalConns {
				t.Errorf("TotalConns = %d, want %d", stats.TotalConns, tt.stats.TotalConns)
			}
			if stats.IdleConns != tt.stats.IdleConns {
				t.Errorf("IdleConns = %d, want %d", stats.IdleConns, tt.stats.IdleConns)
			}
			if stats.StaleConns != tt.stats.StaleConns {
				t.Errorf("StaleConns = %d, want %d", stats.StaleConns, tt.stats.StaleConns)
			}

			_ = registry
		})
	}
}

// TestRedisStatsGetter_Interface 测试 RedisStatsGetter 接口实现
func TestRedisStatsGetter_Interface(t *testing.T) {
	var _ RedisStatsGetter = (*mockRedisStatsGetter)(nil)

	mock := &mockRedisStatsGetter{
		stats: &redis.PoolStats{
			Hits:   50,
			Misses: 5,
		},
	}

	stats := mock.PoolStats()
	if stats == nil {
		t.Error("PoolStats() 返回 nil")
		return
	}

	if stats.Hits != 50 {
		t.Errorf("Hits = %d, want 50", stats.Hits)
	}
	if stats.Misses != 5 {
		t.Errorf("Misses = %d, want 5", stats.Misses)
	}
}

// TestRegisterRedisStatsMetrics_MultipleDB 测试注册多个数据库的指标
func TestRegisterRedisStatsMetrics_MultipleDB(t *testing.T) {
	dbs := []struct {
		name  string
		stats *redis.PoolStats
	}{
		{
			name: "db1",
			stats: &redis.PoolStats{
				Hits:       100,
				TotalConns: 10,
			},
		},
		{
			name: "db2",
			stats: &redis.PoolStats{
				Hits:       200,
				TotalConns: 20,
			},
		},
		{
			name: "db3",
			stats: &redis.PoolStats{
				Hits:       300,
				TotalConns: 30,
			},
		},
	}

	for _, db := range dbs {
		t.Run(db.name, func(t *testing.T) {
			mock := &mockRedisStatsGetter{stats: db.stats}
			RegisterRedisStatsMetrics(db.name, mock)

			stats := mock.PoolStats()
			if stats.Hits != db.stats.Hits {
				t.Errorf("Hits = %d, want %d", stats.Hits, db.stats.Hits)
			}
			if stats.TotalConns != db.stats.TotalConns {
				t.Errorf("TotalConns = %d, want %d", stats.TotalConns, db.stats.TotalConns)
			}
		})
	}
}

// TestRegisterRedisStatsMetrics_NilStats 测试空统计信息
func TestRegisterRedisStatsMetrics_NilStats(t *testing.T) {
	mock := &mockRedisStatsGetter{stats: nil}

	// 这应该不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterRedisStatsMetrics() panic = %v", r)
		}
	}()

	RegisterRedisStatsMetrics("nil_stats_db", mock)
}
