package promdb

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"

	"github.com/spelens-gud/Verktyg/interfaces/imetrics"
)

type dbUsageMetrics struct {
	*imetrics.MetricsGroup
}

func InitDBUsageMetrics(dbType, namespace string) imetrics.DBMetrics {
	label := map[string]string{
		"type": dbType,
	}

	group := imetrics.NewMetricsGroup(namespace, imetrics.SubsystemUsage, label)

	group.NewHistogram(imetrics.NameDurationSeconds, "db client requests duration(s).",
		"name", "addr", "command", "error")

	m := &dbUsageMetrics{MetricsGroup: group}
	imetrics.MustRegister(group)
	return m
}

func (m *dbUsageMetrics) Metrics(name, addr, command string, err error, duration time.Duration) {
	errMsg := ""

	if err != nil {
		switch e := err.(type) {
		case *mysql.MySQLError:
			errMsg = fmt.Sprintf("mysql error: %d", e.Number)
		case redis.Error:
			errMsg = fmt.Sprintf("redis error: %v", e.Error())
		default:
			errMsg = err.Error()
		}
	}

	m.Add(imetrics.NameDurationSeconds, duration.Seconds(), name, addr, command, errMsg)
}
