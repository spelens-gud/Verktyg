package isql

import (
	"context"

	"github.com/jinzhu/gorm"
	gorm2 "gorm.io/gorm"

	"github.com/spelens-gud/Verktyg.git/internal/incontext"
)

type Gorm2SQL interface {
	// 获取gorm连接
	GetDB(ctx context.Context) *gorm2.DB
	// 开启事务
	WithTransaction(ctx context.Context, f func(tx *gorm2.DB) (err error)) (err error)
	// 开启context事务
	Transaction(ctx context.Context, f func(ctx context.Context, tx *gorm2.DB) (err error)) (err error)
	// 关闭连接
	Close() error
}

type GormSQL interface {
	// 获取gorm连接
	GetDB(ctx context.Context) *gorm.DB
	// 开启事务
	// Deprecated: 使用Transaction
	WithTransaction(ctx context.Context, f func(tx *gorm.DB) (err error)) (err error)
	// 开启事务
	Transaction(ctx context.Context, f func(ctx context.Context, tx *gorm.DB) (err error)) (err error)
	// 关闭连接
	Close() error
}

const dbTxKey = incontext.Key("isql.in_context_store")

type dbTxStore map[interface{}]interface{}

func ExtractTransaction(ctx context.Context, dbInstance interface{}) interface{} {
	if store, _ := dbTxKey.Value(ctx).(dbTxStore); len(store) > 0 {
		return store[dbInstance]
	}
	return nil
}

func DetachTransaction(ctx context.Context) context.Context {
	return dbTxKey.WithValue(ctx, nil)
}

func InjectTransaction(ctx context.Context, dbInstance interface{}, tx interface{}) context.Context {
	if store, _ := dbTxKey.Value(ctx).(dbTxStore); len(store) == 0 {
		return dbTxKey.WithValue(ctx, dbTxStore{dbInstance: tx})
	} else {
		newStore := make(dbTxStore, len(store))
		for k, v := range store {
			newStore[k] = v
		}
		newStore[dbInstance] = tx
		return dbTxKey.WithValue(ctx, newStore)
	}
}
