package promdb

import (
	"database/sql"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/spelens-gud/Verktyg.git/interfaces/imetrics"
)

type SqlStatsGetter interface {
	Stats() sql.DBStats
}

func RegisterSqlStatsMetrics(host, port, db string, sg SqlStatsGetter) {
	constLabels := prometheus.Labels{"db_name": strings.Join([]string{host, port, db}, "_")}

	descGroup := imetrics.NewDescGroup(imetrics.NamespaceSqlStats, imetrics.SubsystemConnections, constLabels,
		func() []interface{} {
			return []interface{}{sg.Stats()}
		})

	descGroup.NewConstMetrics("max_open", "Maximum number of open connections to the database.",
		prometheus.GaugeValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).MaxOpenConnections)
		})

	descGroup.NewConstMetrics("open", "The number of established connections both in use and idle.",
		prometheus.GaugeValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).OpenConnections)
		})

	descGroup.NewConstMetrics("in_use", "The number of connections currently in use.",
		prometheus.GaugeValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).InUse)
		})

	descGroup.NewConstMetrics("idle", "The number of idle connections.",
		prometheus.GaugeValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).Idle)
		})

	descGroup.NewConstMetrics("waited_for", "The total number of connections waited for.",
		prometheus.CounterValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).WaitCount)
		})

	descGroup.NewConstMetrics("blocked_seconds", "The total time blocked waiting for a new connection.",
		prometheus.CounterValue, func(items ...interface{}) float64 {
			return items[0].(sql.DBStats).WaitDuration.Seconds()
		})

	descGroup.NewConstMetrics("closed_max_idle", "The total number of connections closed due to SetMaxIdleConns.",
		prometheus.CounterValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).MaxIdleClosed)
		})

	descGroup.NewConstMetrics("closed_max_lifetime", "The total number of connections closed due to SetConnMaxLifetime.",
		prometheus.CounterValue, func(items ...interface{}) float64 {
			return float64(items[0].(sql.DBStats).MaxLifetimeClosed)
		})

	imetrics.MustRegister(descGroup)
}
