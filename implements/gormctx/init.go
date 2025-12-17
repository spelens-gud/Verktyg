package gormctx

import (
	"context"

	"github.com/jinzhu/gorm"

	"github.com/spelens-gud/Verktyg/implements/sqlcx"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/isql"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
)

// 实例化原始gorm.DB
func InitGormDB(dialect string, sqlCommon gorm.SQLCommon, opts ...func(db *gorm.DB)) (gormDb *gorm.DB, err error) {
	gormDb, err = gorm.Open(dialect, sqlCommon)
	if err != nil {
		return nil, err
	}

	if envflag.IsEmpty(iconfig.EnvKeyLogDisableSqlGorm) {
		// 日志开启
		gormDb.LogMode(true)
		// 日志模式
		gormDb.Set("detailMode", !iconfig.GetEnv().IsProduction())
	}

	// 禁用全局update防止漏写where导致严重事故
	gormDb.BlockGlobalUpdate(true)

	// 其他自定义gorm.DB处理
	for _, opt := range opts {
		opt(gormDb)
	}
	return
}

// 单读写sql gorm实例
func NewSimpleGormSql(config isql.SQLConfig, opts ...func(db *gorm.DB)) (db isql.GormSQL, err error) {
	sqlType := "mysql"

	d := sqlcx.GetDriver(sqlType)
	// 实例化写连接
	conn, err := sqlcx.InitSqlDbConn(&config, d)
	if err != nil {
		return nil, err
	}

	gormDb, err := InitGormDB(sqlType, &ctxConn{
		write:   conn,
		read:    conn,
		context: context.Background(),
	}, opts...)
	if err != nil {
		return
	}

	replaceTransactionHook(gormDb)

	db = &wdb{
		write:  conn,
		gormDB: gormDb,
	}
	return
}

// 一写多读sql gorm实例
func NewGormSql(config isql.WRSQLConfig, opts ...func(db *gorm.DB)) (db isql.GormSQL, err error) {
	sqlType := "mysql"

	d := sqlcx.GetDriver(sqlType)
	// 实例化写连接
	writeConn, err := sqlcx.InitSqlDbConn(&config.Write, d)
	if err != nil {
		return nil, err
	}

	// 实例化读连接
	readConns := make([]connDB, 0, len(config.Reads))
	for _, readConfig := range config.Reads {
		read, err := sqlcx.InitSqlDbConn(&readConfig, d)
		if err != nil {
			return nil, err
		}
		readConns = append(readConns, read)
	}

	gormDb, err := InitGormDB(sqlType, &ctxConn{
		write:   writeConn,
		read:    writeConn,
		context: context.Background(),
	}, opts...)
	if err != nil {
		return
	}

	replaceTransactionHook(gormDb)

	mainDB := wdb{
		write:  writeConn,
		gormDB: gormDb,
	}

	if len(readConns) > 0 {
		db = &wrdb{
			wdb:  mainDB,
			read: readConns,
		}
	} else {
		db = &mainDB
	}
	return
}

// DEPRECATED: 不要再用这种写法 请使用类型别名
func NewMultiGormSQL(cfg isql.MultiWRSQLConfig, multiDB interface{}, opts ...func(db *gorm.DB)) (cf func(), err error) {
	return isql.NewMultiSqlFactory(func(config isql.WRSQLConfig) (isql.GormSQL, error) {
		return NewGormSql(config, opts...)
	})(cfg, multiDB)
}
