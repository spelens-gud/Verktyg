package gormctx

import (
	"context"
	"database/sql"
	"io"

	"github.com/jinzhu/gorm"
)

var (
	gormDBOffset  uintptr
	ctxBackground = context.Background()

	_ sqlTx          = &ctxTxConn{}
	_ gorm.SQLCommon = &ctxTxConn{}
	_ sqlTxDb        = &ctxConn{}
	_ gorm.SQLCommon = &ctxConn{}
)

type (
	sqlTxDb interface {
		Begin() (*sql.Tx, error)
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	}

	sqlTx interface {
		Commit() error
		Rollback() error
	}

	contextCommonDB interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}

	connDB interface {
		contextCommonDB
		sqlTxDb
		io.Closer
	}

	ctxConn struct {
		write   connDB
		read    connDB
		context context.Context
	}

	ctxTxConn struct {
		*sql.Tx
		context context.Context
	}
)

func (conn *ctxConn) BeginContextTx(ctx context.Context, options *sql.TxOptions) (gorm.SQLCommon, error) {
	tx, err := conn.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &ctxTxConn{context: ctx, Tx: tx}, err
}

// 连接封装
func (conn *ctxConn) Begin() (*sql.Tx, error) {
	return conn.BeginTx(conn.context, nil)
}

func (conn *ctxConn) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	if ctx == ctxBackground || ctx == nil {
		ctx = conn.context
	}
	return conn.write.BeginTx(ctx, options)
}

func (conn *ctxConn) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	return conn.write.ExecContext(conn.context, query, args...)
}

func (conn *ctxConn) Prepare(query string) (stmt *sql.Stmt, err error) {
	return conn.write.PrepareContext(conn.context, query)
}

func (conn *ctxConn) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return conn.read.QueryContext(conn.context, query, args...)
}

func (conn *ctxConn) QueryRow(query string, args ...interface{}) *sql.Row {
	return conn.read.QueryRowContext(conn.context, query, args...)
}

// 事务连接封装
func (conn *ctxTxConn) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	return conn.ExecContext(conn.context, query, args...)
}

func (conn *ctxTxConn) Prepare(query string) (stmt *sql.Stmt, err error) {
	return conn.PrepareContext(conn.context, query)
}

func (conn *ctxTxConn) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return conn.QueryContext(conn.context, query, args...)
}

func (conn *ctxTxConn) QueryRow(query string, args ...interface{}) *sql.Row {
	return conn.QueryRowContext(conn.context, query, args...)
}
