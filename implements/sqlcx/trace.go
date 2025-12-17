package sqlcx

import (
	"context"
	"database/sql/driver"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/spelens-gud/Verktyg.git/interfaces/isql"
	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/kdb"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
	"github.com/spelens-gud/Verktyg.git/kits/ktrace/tracer"
)

// trace 追踪
func startExitSpan(ctx context.Context, config *isql.SQLConfig, operation, statement string) (span itrace.Span) {
	span = tracer.StartExitSpan(ctx, "MYSQL_CLIENT:"+operation, config.Host+":"+strconv.Itoa(config.Port))

	span.SetComponent(itrace.ComponentMysql)
	span.SetLayer(itrace.SpanLayerDatabase)

	span.Tag(string(ext.DBStatement), statement)
	span.Tag(itrace.TagDBStatementType, operation)
	span.Tag(string(ext.DBInstance), config.DB)
	span.Tag(string(ext.DBType), "mysql")
	return
}

func startProgressSpan(ctx context.Context, config *isql.SQLConfig, operation, statement string) (span itrace.Span, nCtx context.Context) {
	span, nCtx = tracer.StartSpan(ctx, "MYSQL_CLIENT:"+operation)

	span.SetComponent(itrace.ComponentMysql)
	span.SetLayer(itrace.SpanLayerDatabase)

	span.Tag(string(ext.DBStatement), statement)
	span.Tag(itrace.TagDBStatementType, operation)
	span.Tag(string(ext.DBInstance), config.DB)
	span.Tag(string(ext.DBType), "mysql")
	return
}

type SqlTelemetryReporter struct {
	span               itrace.Span
	tStart             time.Time
	config             *isql.SQLConfig
	ctx                context.Context
	operation, command string
	disableErrorLog    bool
}

func (t *SqlTelemetryReporter) Report(err error) {
	if err == driver.ErrSkip {
		err = nil
	}

	if strings.Contains(t.operation, "Close") && err == driver.ErrBadConn {
		err = nil
	}

	var (
		duration = time.Since(t.tStart)
		ctx      = tracer.SpanWithContext(t.ctx, t.span)
	)

	// 指标统计
	kdb.SqlMetrics(t.operation, t.config.Host, "", err, duration)
	// 错误日志 慢日志
	t.doLog(ctx, duration, err)
	// 链路追踪
	if err != nil {
		t.span.Error(err, logger.FromContext(ctx).FieldData())
	}
	t.span.Finish()
}

// 慢日志 和 错误日志记录
func (t *SqlTelemetryReporter) doLog(ctx context.Context, du time.Duration, err error) {
	// 基准标签
	fields := map[string]interface{}{
		"sql":       t.command,
		"operation": t.operation,
		"addr":      t.config.Host,
		"duration":  du.String(),
		"cost":      du.Seconds(),
	}
	if len(t.command) > 0 {
		fields["sql"] = t.command
	}

	lg := logger.FromContext(ctx).WithTag("MYSQL").WithFields(fields)

	// 慢日志
	if t.config.SlowLogThresholdMs > 0 && du > time.Duration(t.config.SlowLogThresholdMs)*time.Millisecond {
		lg.WithTag("MYSQL_SLOW").Warn("mysql slow log")
	}

	// 错误日志
	if err != nil && !t.disableErrorLog {
		errMsg := err.Error()
		if err != context.Canceled && !strings.Contains(errMsg, "dial tcp: operation was canceled") {
			lg = lg.WithTag("MYSQL_ERR")
		}
		lg.WithField("err", errMsg).Error("mysql err log")
	}
}
