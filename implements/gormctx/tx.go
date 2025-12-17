package gormctx

import (
	"context"
	"database/sql"

	"github.com/jinzhu/gorm"
	"go.uber.org/multierr"

	"github.com/spelens-gud/Verktyg.git/interfaces/isql"
)

type GormContextTx interface {
	BeginContextTx(ctx context.Context, options *sql.TxOptions) (db gorm.SQLCommon, err error)
}

func withTransaction(ctx context.Context, db isql.GormSQL, f func(ctx context.Context, tx *gorm.DB) error) (err error) {
	if tx, ok := isql.ExtractTransaction(ctx, db).(*gorm.DB); ok {
		return f(ctx, tx)
	}

	var (
		txDB         = db.GetDB(ctx) // 实例化新写连接
		dbCommon, ok = txDB.CommonDB().(GormContextTx)
	)

	if !ok {
		// 某些原因导致执行器替换没有成功
		return txDB.Transaction(func(tx *gorm.DB) error {
			return f(isql.InjectTransaction(ctx, db, tx), tx)
		})
	}

	// 开启事务连接
	txCommonDB, err := dbCommon.BeginContextTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	// 更新执行db
	setGormCommonDB(txDB, txCommonDB)

	var (
		errs    []error
		success bool // 调用f时如果出现panic，err则会无法正常赋值，因此需要此变量

	)

	defer func() {
		// 不成功
		if !success {
			// 回滚事务
			if e := txDB.Rollback().Error; e != nil {
				errs = append(errs, e)
			}
		}
		err = multierr.Combine(errs...)
	}()

	// 执行sql闭包方法
	if e := f(isql.InjectTransaction(ctx, db, txDB), txDB); e != nil {
		errs = append(errs, e)
		return
	}

	// 如果是新创建的事务 则尝试提交
	// 否则是在事务中 不应该由此层提交
	if e := txDB.Commit().Error; e != nil {
		errs = append(errs, e)
		return
	}

	// 赋值success
	success = true
	return
}

func beginTx(scope *gorm.Scope) {
	if wrapDB, ok := scope.SQLDB().(*ctxConn); ok {
		if tx, err := wrapDB.BeginContextTx(wrapDB.context, &sql.TxOptions{}); scope.Err(err) == nil {
			setGormCommonDB(scope.DB(), tx)
			scope.InstanceSet("gorm:started_transaction", true)
		}
	}
}
