package isql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const defaultMaxInsertRows = 1000

// Row ...
type (
	Row []interface{}

	batchInsertOption struct {
		MaxInsertRows int
		IsIgnore      bool   // insert ignore
		OnDuplicate   string // insert into ... on duplicate update...
	}

	BatchInsertOption func(o *batchInsertOption)
)

func OnDuplicate(fields ...string) BatchInsertOption {
	return func(o *batchInsertOption) {
		for i := range fields {
			fields[i] = "`" + fields[i] + "`= VALUES(`" + fields[i] + "`)"
		}
		o.OnDuplicate = "ON DUPLICATE KEY UPDATE " + strings.Join(fields, ",")
	}
}

func IsIgnore() BatchInsertOption {
	return func(o *batchInsertOption) {
		o.IsIgnore = true
	}
}

func parentheses(s string) string {
	return "(" + s + ")"
}

// InsertBatch 批量插入数据
func InsertBatch(ctx context.Context, db SQLCommand, table string, columns []string, rows []Row, opts ...BatchInsertOption) error {
	opt := new(batchInsertOption)
	for _, o := range opts {
		o(opt)
	}
	if db == nil {
		panic("invalid db")
	}
	if table == "" || len(columns) == 0 || len(rows) == 0 {
		return nil
	}
	maxInsertRows := opt.MaxInsertRows
	if maxInsertRows <= 0 {
		maxInsertRows = defaultMaxInsertRows
	}
	if len(rows) <= maxInsertRows {
		return insertBatch(ctx, db, table, columns, rows, opt)
	} else {
		return db.WithTransaction(ctx, func(tx SQLCommand) (err error) {
			rr := len(rows) / maxInsertRows
			for i := 0; i < rr+1; i++ {
				l := (i + 1) * maxInsertRows
				if l > len(rows) {
					l = len(rows)
				}
				if i*maxInsertRows < l {
					err = insertBatch(ctx, tx, table, columns, rows[i*maxInsertRows:l], opt)
					if err != nil {
						return
					}
				}
			}
			return
		})
	}
}

func BatchInsert(ctx context.Context, db SQLCommand, table string, rowsIn interface{}, rowsMethod func(i int, insert func(...interface{})), columns []string, options ...BatchInsertOption) (err error) {
	value := reflect.ValueOf(rowsIn)
	if value.Kind() != reflect.Slice {
		err = errors.New("invalid slice")
		return
	}
	l := value.Len()
	if l == 0 {
		return
	}
	rows := make([]Row, 0, l)
	for i := 0; i < l; i++ {
		rowsMethod(i, func(in ...interface{}) {
			rows = append(rows, in)
		})
	}
	return InsertBatch(ctx, db, table, columns, rows, options...)
}

// 分批插入 如果db在事务中则使用原事务 否则新建事务
func insertBatch(ctx context.Context, db SQLCommand, table string, columns []string, rows []Row, opt *batchInsertOption) error {
	// 格式
	var format string
	if opt.IsIgnore {
		format = " INSERT IGNORE INTO %s(%s) VALUES %s "
	} else {
		format = " INSERT INTO %s(%s) VALUES %s "
	}
	const placeholder = "?"
	const placeholderSep = ","

	// 列数量
	colNum := len(columns)

	// 占位符
	var placeholderList = make([]string, colNum)
	for i := 0; i < colNum; i++ {
		if !strings.HasPrefix(columns[i], "`") {
			columns[i] = "`" + columns[i] + "`"
		}
		placeholderList[i] = placeholder
	}
	rowPlaceholder := parentheses(strings.Join(placeholderList, placeholderSep))

	// 行
	var values []string
	var args []interface{}
	for _, row := range rows {
		if len(row) != colNum {
			return fmt.Errorf("row (%v) does not match columns (%v)", row, columns)
		}
		values = append(values, rowPlaceholder)
		args = append(args, row...)
	}

	// 拼接sql并执行
	insertSql := fmt.Sprintf(format, table, strings.Join(columns, placeholderSep), strings.Join(values, placeholderSep))
	if opt.OnDuplicate != "" {
		insertSql += opt.OnDuplicate
	}
	if err := db.Exec(ctx, insertSql, args...); err != nil {
		return err
	}
	return nil
}
