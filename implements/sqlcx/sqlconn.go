package sqlcx

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/ngrok/sqlmw"

	"github.com/spelens-gud/Verktyg/implements/promdb"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/isql"
	"github.com/spelens-gud/Verktyg/kits/kdb"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
)

func GetDriver(name string) (d driver.Driver) {
	switch name {
	case "mysql":
		return &mysql.MySQLDriver{}
	case "postgres":
		return &pq.Driver{}
	}
	return &mysql.MySQLDriver{}
}

var disableErrorLog = envflag.BoolOnceFrom(iconfig.EnvKeyLogDisableSqlError)

func InitSqlDbConn(config *isql.SQLConfig, d driver.Driver) (*sql.DB, error) {
	// 配置初始化
	config.Init()

	if !config.DisableInterceptor {
		// 注入拦截器
		d = sqlmw.Driver(d, Interceptor{
			config:          config,
			disableErrorLog: disableErrorLog(),
		})
	}

	conn, err := d.(driver.DriverContext).OpenConnector(config.ConnectionUrl())
	if err != nil {
		return nil, err
	}

	// 实例化连接池
	sqlDB := sql.OpenDB(conn)

	// 连接池配置
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifeTimeMs) * time.Millisecond)

	// 兼容旧版本
	if idleTimeSet, ok := (interface{})(sqlDB).(interface {
		SetConnMaxIdleTime(duration time.Duration)
	}); ok {
		idleTimeSet.SetConnMaxIdleTime(time.Duration(config.MaxIdleTimeMs) * time.Millisecond)
	}

	// 检查连接
	if err := kdb.PingDB(sqlDB.Ping, "sql", fmt.Sprint(config.Host, ":", config.Port)); err != nil {
		return nil, err
	}

	// 注册埋点监控
	promdb.RegisterSqlStatsMetrics(config.Host, strconv.Itoa(config.Port), config.DB, sqlDB)
	return sqlDB, nil
}
