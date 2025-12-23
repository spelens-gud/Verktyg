package promstore

import (
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/imetrics"
)

type storeMetrics struct {
	*imetrics.MetricsGroup
}

func NewStoreMetrics(name string) *storeMetrics {
	group := imetrics.NewMetricsGroup("cache", "store", map[string]string{
		"name": name,
	})

	group.NewHistogram(imetrics.NameHandlingSeconds, "Histogram of response latency (seconds) of store action.",
		"action", "error",
	)
	m := &storeMetrics{group}
	imetrics.MustRegister(m)
	return m
}

func (m *storeMetrics) Observe(action string, err error, duration time.Duration) {
	isErr := ""
	if err != nil {
		isErr = "error"
	}
	m.Add(imetrics.NameHandlingSeconds, duration.Seconds(), action, isErr)
}
