package prommq

import (
	"time"

	"github.com/spelens-gud/Verktyg.git/interfaces/imetrics"
)

type mqUsageMetrics struct {
	*imetrics.MetricsGroup
}

func InitMQUsageMetrics(mqType string) imetrics.MqMetrics {
	group := imetrics.NewMetricsGroup(imetrics.NamespaceMq, imetrics.SubsystemClient, map[string]string{
		"type": mqType,
	})

	group.NewHistogram(imetrics.NameDurationSeconds, "mq client process duration(s).",
		"addr", "action", "topic", "error",
	)

	m := &mqUsageMetrics{group}
	imetrics.MustRegister(m)
	return m
}

func (m *mqUsageMetrics) MetricsProduce(addr, topic string, err error, duration time.Duration) {
	m.metrics(addr, "produce", topic, err, duration)
}

func (m *mqUsageMetrics) MetricsConsume(addr, topic string, err error, duration time.Duration) {
	m.metrics(addr, "consume", topic, err, duration)
}

func (m *mqUsageMetrics) metrics(addr, action, topic string, err error, duration time.Duration) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	m.Add(imetrics.NameDurationSeconds, duration.Seconds(),
		addr, action, topic, errMsg)
}
