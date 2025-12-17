package gormctx

import (
	"context"
	"database/sql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/spelens-gud/Verktyg/implements/sqlcx"
	"github.com/spelens-gud/Verktyg/interfaces/isql"
)

var _ isql.Gorm2SQL = &gorm2Sql{}

type gorm2Sql struct {
	DB   *gorm.DB
	conn *sql.DB
}

func (g *gorm2Sql) Transaction(ctx context.Context, f func(ctx context.Context, tx *gorm.DB) (err error)) (err error) {
	if tx, ok := isql.ExtractTransaction(ctx, g).(*gorm.DB); ok {
		return f(ctx, tx)
	}
	return g.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = isql.InjectTransaction(ctx, g, tx)
		return f(ctx, tx.WithContext(ctx))
	})
}

func (g *gorm2Sql) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := isql.ExtractTransaction(ctx, g).(*gorm.DB); ok {
		return tx
	}
	return g.DB.WithContext(ctx)
}

func (g *gorm2Sql) WithTransaction(ctx context.Context, f func(tx *gorm.DB) (err error)) (err error) {
	return g.Transaction(ctx, func(_ context.Context, tx *gorm.DB) (err error) {
		return f(tx)
	})
}

func (g *gorm2Sql) Close() error {
	return g.conn.Close()
}

func InitGormV2Sql(config isql.SQLConfig, gormOpts ...func(*gorm.Config)) (sql isql.Gorm2SQL, err error) {
	sqlType := "mysql"
	db, err := sqlcx.InitSqlDbConn(&config, sqlcx.GetDriver(sqlType))
	if err != nil {
		return
	}

	gormCfg := &gorm.Config{
		Logger: newGorm2Logger(),
	}
	for _, o := range gormOpts {
		o(gormCfg)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), gormCfg)
	if err != nil {
		return
	}

	sql = &gorm2Sql{DB: gormDB, conn: db}
	return
}
