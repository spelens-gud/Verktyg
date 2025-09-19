package sqlcx

import (
	"context"
	"database/sql/driver"
	"strconv"

	"git.bestfulfill.tech/devops/go-core/interfaces/itrace"
)

type wRows struct {
	driver.Rows
	querySpan itrace.Span
	rowSpan   itrace.Span
	statement string
	count     int
	ctx       context.Context
}

var _ wrapRows = &wRows{}

type wrapRows interface {
	Statement() string
	End()
	Context() context.Context
	Incr()
}

func (r *wRows) Context() context.Context {
	return r.ctx
}

func (r *wRows) End() {
	if r.rowSpan != nil {
		r.rowSpan.Tag("db.row_count", strconv.Itoa(r.count))
		r.rowSpan.Finish()
	}
	if r.querySpan != nil {
		r.querySpan.Finish()
	}
}

func (r *wRows) Incr() {
	r.count += 1
}

func (r *wRows) Statement() string {
	return r.statement
}

type wTx struct {
	driver.Tx
	ctx    context.Context
	txSpan itrace.Span
}

func (w *wTx) End() {
	if w.txSpan != nil {
		w.txSpan.Finish()
	}
}

func (w *wTx) Context() context.Context {
	return w.ctx
}

var _ wrapTx = &wTx{}

type wrapTx interface {
	End()
	Context() context.Context
}
