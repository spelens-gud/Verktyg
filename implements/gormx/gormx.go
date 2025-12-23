package gormx

import (
	"context"
	"database/sql"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/spelens-gud/Verktyg/interfaces/isql"
)

type sqlDb interface {
	Begin() (*sql.Tx, error)
}

var _ isql.SQLCommand = &Gormx{}

type Gormx struct {
	db      *gorm.DB
	selects []string
}

func NewWithGormDB(db *gorm.DB) isql.SQLCommand {
	return &Gormx{
		db:      db,
		selects: nil,
	}
}

func (db *Gormx) Close() error {
	return db.db.Close()
}

func (db *Gormx) New() isql.SQLCommand {
	return &Gormx{
		db: db.db.New(),
	}
}

func (db *Gormx) DB() interface{} {
	return db.db
}

func canStartTransaction(db *Gormx) (is bool) {
	sqlDb, ok := db.db.CommonDB().(sqlDb)
	return ok && sqlDb != nil
}

func (db *Gormx) WithTransaction(ctx context.Context, f func(tx isql.SQLCommand) (err error)) error {
	var tx *gorm.DB
	var isNewTx bool
	if canStartTransaction(db) {
		tx = db.db.BeginTx(ctx, &sql.TxOptions{})
		isNewTx = true
	} else {
		tx = db.db
	}

	var err error
	var success bool // 调用f时如果出现panic，err则会无法正常赋值，因此需要此变量
	defer func() {
		if !success {
			if isNewTx {
				tx.Rollback() // 执行f时出现任何问题，都要Rollback
			}
		}
	}()

	err = f(&Gormx{db: tx})
	if err == nil {
		success = true
		if isNewTx {
			return tx.Commit().Error // 成功则提交
		}
	}
	return err
}

func (db *Gormx) Order(order string) isql.SQLCommand {
	db.db = db.db.Order(order)
	return db
}

func (db *Gormx) Offset(offset int) isql.SQLCommand {
	db.db = db.db.Offset(offset)
	return db
}

func (db *Gormx) Limit(limit int) isql.SQLCommand {
	db.db = db.db.Limit(limit)
	return db
}

func (db *Gormx) ScanOffset(ctx context.Context, dst interface{}, order string, offset, limit int) (err error) {
	return db.db.Offset(offset).Limit(limit).Order(order).Scan(dst).Error
}

func (db *Gormx) UseWriteDB() isql.SQLCommand {
	panic("implement me")
}

func (db *Gormx) UseReadDB() isql.SQLCommand {
	panic("implement me")
}

func (db *Gormx) Table(table, tag string) isql.SQLCommand {
	if table == "" {
		return db
	}
	if tag != "" {
		db.db = db.db.Table("`" + table + "` AS " + tag)
	} else {
		db.db = db.db.Table(table)
	}
	return db
}

func (db *Gormx) Where(where string, args ...interface{}) isql.SQLCommand {
	db.db = db.db.Where(where, args...)
	return db
}

func (db *Gormx) match(field string, value interface{}, match bool) isql.SQLCommand {
	if field == "" {
		return db
	}
	m := " = ?"
	if !match {
		m = " != ?"
	}
	db.db = db.db.Where(field+m, value)
	return db
}

// where field = value
func (db *Gormx) WhereEq(field string, value interface{}) isql.SQLCommand {
	return db.match(field, value, true)
}

// where field != value
func (db *Gormx) WhereNeq(field string, value interface{}) isql.SQLCommand {
	return db.match(field, value, false)
}

func (db *Gormx) WhereIf(state bool, where string, args ...interface{}) isql.SQLCommand {
	if state {
		return db.Where(where, args)
	} else {
		return db
	}
}

func (db *Gormx) Join(typ string, table, as, on string, args ...interface{}) isql.SQLCommand {
	if table == "" || on == "" {
		return db
	}
	if as == "" {
		as = table
	}
	join := typ + " JOIN `" + table + "` AS " + as + " ON " + on
	db.db = db.db.Joins(join, args...)
	return db
}

func (db *Gormx) LeftJoin(table, as, on string, args ...interface{}) isql.SQLCommand {
	return db.Join("LEFT", table, as, on, args...)
}

func (db *Gormx) RightJoin(table, as, on string, args ...interface{}) isql.SQLCommand {
	return db.Join("RIGHT", table, as, on, args...)
}

func (db *Gormx) InnerJoin(table, as, on string, args ...interface{}) isql.SQLCommand {
	return db.Join("INNER", table, as, on, args...)
}

func (db *Gormx) OuterJoin(table, as, on string, args ...interface{}) isql.SQLCommand {
	return db.Join("OUTER", table, as, on, args...)
}

func (db *Gormx) Select(fields ...string) isql.SQLCommand {
	db.selects = append(db.selects, fields...)
	return db
}

func (db *Gormx) Exec(ctx context.Context, sql string, args ...interface{}) (err error) {
	return db.db.Exec(sql, args...).Error
}

func (db *Gormx) PageQuery(ctx context.Context, pageParam isql.PageParam, dst interface{}) (total int, err error) {
	countDB := db.db
	if len(pageParam.Distinct) > 0 {
		countDB = db.db.Select(pageParam.Distinct)
	}
	return isql.GetPageResult(ctx, isql.PageSQLParam{
		CountQuery: &Gormx{
			db: countDB,
		},
		DataQuery: db,
		Order:     pageParam.Order,
		PageNo:    pageParam.No,
		PageSize:  pageParam.Count,
	}, dst)
}

func (db *Gormx) Update(ctx context.Context, dst interface{}, where string, arg ...interface{}) (err error) {
	return db.db.Where(where, arg...).UpdateColumn(dst).Error
}

func (db *Gormx) Create(ctx context.Context, dst interface{}) (err error) {
	err = db.db.Create(dst).Error
	return
}

func (db *Gormx) Delete(ctx context.Context, table, where string, args ...interface{}) (err error) {
	// todo:更好的调用方式
	err = db.db.Delete(&table, where, args).Error
	return
}

func (db *Gormx) Scan(ctx context.Context, dst interface{}) (err error) {
	if len(db.selects) > 0 {
		db.db = db.db.Select(strings.Join(db.selects, ","))
	}
	return db.db.Scan(dst).Error
}

// 如果 value 非空 则 添加  where field = value
func (db *Gormx) WhereEqIfNotNull(field string, value interface{}) isql.SQLCommand {
	switch s := value.(type) {
	case int:
		if s != 0 {
			return db.WhereEq(field, value)
		}
	case int8:
		if s != int8(0) {
			return db.WhereEq(field, value)
		}
	case int16:
		if s != int16(0) {
			return db.WhereEq(field, value)
		}
	case int32:
		if s != int32(0) {
			return db.WhereEq(field, value)
		}
	case int64:
		if s != int64(0) {
			return db.WhereEq(field, value)
		}
	case uint:
		if s != uint(0) {
			return db.WhereEq(field, value)
		}
	case uint8:
		if s != uint8(0) {
			return db.WhereEq(field, value)
		}
	case uint16:
		if s != uint16(0) {
			return db.WhereEq(field, value)
		}
	case uint32:
		if s != uint32(0) {
			return db.WhereEq(field, value)
		}
	case uint64:
		if s != uint64(0) {
			return db.WhereEq(field, value)
		}
	case uintptr:
		if s != uintptr(0) {
			return db.WhereEq(field, value)
		}
	case float32:
		if s != float32(0) {
			return db.WhereEq(field, value)
		}
	case float64:
		if s != float64(0) {
			return db.WhereEq(field, value)
		}
	case string:
		if s != "" {
			return db.WhereEq(field, value)
		}
	default:
		k := reflect.ValueOf(s)
		if !k.IsNil() {
			return db.WhereEq(field, value)
		}
	}
	return db
}

func (db *Gormx) Count(ctx context.Context) (count int, err error) {
	err = db.db.Count(&count).Error
	return
}

func (db *Gormx) BatchInsert(ctx context.Context, table string, rowsIn interface{}, rowsMethod isql.BatchInsertFunc, columns []string, option ...isql.BatchInsertOption) (err error) {
	return isql.BatchInsert(ctx, db, table, rowsIn, rowsMethod, columns, option...)
}
