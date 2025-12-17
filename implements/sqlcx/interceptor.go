package sqlcx

import (
	"context"
	"database/sql/driver"
	"io"
	"time"

	"github.com/ngrok/sqlmw"

	"github.com/spelens-gud/Verktyg/interfaces/isql"
	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

type Interceptor struct {
	config *isql.SQLConfig
	sqlmw.NullInterceptor
	disableErrorLog bool
}

var ignoreTraceOperate = map[string]bool{
	"driver:RowsNext": true,
}

func (i *Interceptor) newSqlTelemetryReporter(ctx context.Context, operation, command string) *SqlTelemetryReporter {
	r := &SqlTelemetryReporter{
		config:          i.config,
		ctx:             ctx,
		operation:       operation,
		command:         command,
		tStart:          time.Now(),
		disableErrorLog: i.disableErrorLog,
	}

	if !ignoreTraceOperate[r.operation] {
		r.span = startExitSpan(ctx, i.config, operation, command)
	} else {
		r.span = itrace.NoopSpan
	}
	return r
}

func (i Interceptor) ConnPing(ctx context.Context, pinger driver.Pinger) error {
	reporter := i.newSqlTelemetryReporter(ctx, "driver:ConnPing", "")
	err := pinger.Ping(ctx)
	reporter.Report(err)
	return err
}

func (i Interceptor) ConnectorConnect(ctx context.Context, connector driver.Connector) (driver.Conn, error) {
	reporter := i.newSqlTelemetryReporter(ctx, "driver:ConnectorConnect", "")
	ret, err := connector.Connect(ctx)
	reporter.Report(err)
	return ret, err
}

func (i Interceptor) ConnBeginTx(ctx context.Context, tx driver.ConnBeginTx, options driver.TxOptions) (driver.Tx, error) {
	txSpan, ctx := startProgressSpan(ctx, i.config, "interceptor:Transaction", "")

	reporter := i.newSqlTelemetryReporter(ctx, "driver:ConnBeginTx", "")
	ret, err := tx.BeginTx(ctx, options)
	reporter.Report(err)

	if ret != nil {
		ret = &wTx{
			Tx:     ret,
			ctx:    ctx,
			txSpan: txSpan,
		}
	} else {
		txSpan.Error(err, nil)
		txSpan.Finish()
	}

	return ret, err
}

func (i Interceptor) TxCommit(ctx context.Context, tx driver.Tx) error {
	if w, ok := tx.(wrapTx); ok {
		ctx = w.Context()
		defer w.End()
	}

	reporter := i.newSqlTelemetryReporter(ctx, "driver:TxCommit", "")
	err := tx.Commit()
	reporter.Report(err)
	return err
}

func (i Interceptor) TxRollback(ctx context.Context, tx driver.Tx) error {
	if w, ok := tx.(wrapTx); ok {
		defer w.End()
		ctx = w.Context()
	}

	reporter := i.newSqlTelemetryReporter(ctx, "driver:TxRollback", "")
	err := tx.Rollback()
	reporter.Report(err)
	return err
}

func (i Interceptor) ConnExecContext(ctx context.Context, context driver.ExecerContext, s string, values []driver.NamedValue) (driver.Result, error) {
	reporter := i.newSqlTelemetryReporter(ctx, "driver:ConnExec", s)
	ret, err := context.ExecContext(ctx, s, values)
	reporter.Report(err)
	return ret, err
}

func (i Interceptor) ConnQueryContext(ctx context.Context, context driver.QueryerContext, s string, values []driver.NamedValue) (driver.Rows, error) {
	querySpan, ctx := startProgressSpan(ctx, i.config, "interceptor:ConnQueryRows", s)

	reporter := i.newSqlTelemetryReporter(ctx, "driver:ConnQuery", s)
	ret, err := context.QueryContext(ctx, s, values)
	reporter.Report(err)

	if ret != nil {
		rowSpan, nCtx := startProgressSpan(ctx, i.config, "driver:RowsScan", s)
		ret = &wRows{
			Rows:      ret,
			statement: s,
			rowSpan:   rowSpan,
			querySpan: querySpan,
			ctx:       nCtx,
		}
	} else {
		if err != driver.ErrSkip {
			querySpan.Error(err, nil)
		} else {
			querySpan.Tag("db.driver_skip", driver.ErrSkip.Error())
		}
		querySpan.Finish()
	}

	return ret, err
}

func (i Interceptor) RowsNext(ctx context.Context, rows driver.Rows, values []driver.Value) error {
	// 提取sql语句
	command := ""
	sr, ok := rows.(wrapRows)
	if ok {
		command = sr.Statement()
		ctx = sr.Context()
	}

	reporter := i.newSqlTelemetryReporter(ctx, "driver:RowsNext", command)
	err := rows.Next(values)

	if err == io.EOF {
		reporter.Report(nil)
	} else {
		if err == nil && ok {
			sr.Incr()
		}
		reporter.Report(err)
	}
	return err
}

func (i Interceptor) RowsClose(ctx context.Context, rows driver.Rows) error {
	// 提取sql语句
	command := ""
	sr, ok := rows.(wrapRows)
	if ok {
		command = sr.Statement()
		defer sr.End()
		ctx = sr.Context()
	}

	reporter := i.newSqlTelemetryReporter(ctx, "driver:RowsClose", command)
	err := rows.Close()
	reporter.Report(err)
	return err
}

func (i Interceptor) ConnPrepareContext(ctx context.Context, context driver.ConnPrepareContext, s string) (driver.Stmt, error) {
	reporter := i.newSqlTelemetryReporter(ctx, "driver:ConnPrepareStmt", s)
	ret, err := context.PrepareContext(ctx, s)
	reporter.Report(err)
	return ret, err
}

func (i Interceptor) StmtExecContext(ctx context.Context, context driver.StmtExecContext, s string, values []driver.NamedValue) (driver.Result, error) {
	reporter := i.newSqlTelemetryReporter(ctx, "driver:StmtExec", s)
	ret, err := context.ExecContext(ctx, values)
	reporter.Report(err)
	return ret, err
}

func (i Interceptor) StmtQueryContext(ctx context.Context, context driver.StmtQueryContext, s string, values []driver.NamedValue) (driver.Rows, error) {
	querySpan, ctx := startProgressSpan(ctx, i.config, "interceptor:StmtQueryRows", s)

	reporter := i.newSqlTelemetryReporter(ctx, "driver:StmtQuery", s)
	ret, err := context.QueryContext(ctx, values)
	reporter.Report(err)

	if ret != nil {
		rowSpan, nCtx := startProgressSpan(ctx, i.config, "driver:RowsScan", s)
		ret = &wRows{
			Rows:      ret,
			statement: s,
			rowSpan:   rowSpan,
			querySpan: querySpan,
			ctx:       nCtx,
		}
	} else {
		querySpan.Error(err, nil)
		querySpan.Finish()
	}
	return ret, err
}

func (i Interceptor) StmtClose(ctx context.Context, stmt driver.Stmt) error {
	reporter := i.newSqlTelemetryReporter(ctx, "driver:StmtClose", "")
	err := stmt.Close()
	reporter.Report(err)
	return err
}

var _ sqlmw.Interceptor = &Interceptor{}
