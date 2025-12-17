package isql

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/spelens-gud/Verktyg/kits/knet"
)

type (
	SQLCommand interface {
		// 链接相关

		// 新建链接
		New() SQLCommand
		// 返回实际的DB
		DB() interface{}
		// 开启事务
		WithTransaction(ctx context.Context, f func(tx SQLCommand) (err error)) error
		// 关闭连接
		Close() error

		// 选项命令
		UseWriteDB() SQLCommand
		UseReadDB() SQLCommand

		// 查询命令
		Table(where, as string) SQLCommand
		// where
		Where(where string, args ...interface{}) SQLCommand

		// 如果state成立 则添加where条件
		// if state{
		//    sql = sql.Where(where,args...)
		// }
		WhereIf(state bool, where string, args ...interface{}) SQLCommand

		// 如果参数非零值 则添加字段where比较条件
		// if !isZeroValue(param){
		//    sql = sql.Where(field+" = ?",param)
		// }
		WhereEqIfNotNull(field string, param interface{}) SQLCommand

		// 连表命令
		LeftJoin(table, as, on string, args ...interface{}) SQLCommand
		RightJoin(table, as, on string, args ...interface{}) SQLCommand
		InnerJoin(table, as, on string, args ...interface{}) SQLCommand
		OuterJoin(table, as, on string, args ...interface{}) SQLCommand

		// 选择命令
		Select(fields ...string) SQLCommand

		// 最终执行命令

		// 执行
		Exec(ctx context.Context, sql string, args ...interface{}) (err error)
		// 分页查询
		PageQuery(ctx context.Context, pageParam PageParam, dst interface{}) (total int, err error)
		// 更新
		Update(ctx context.Context, dst interface{}, where string, arg ...interface{}) (err error)
		// 新建
		Create(ctx context.Context, dst interface{}) (err error)
		// 删除
		Delete(ctx context.Context, table, where string, args ...interface{}) (err error)
		// 普通scan
		Scan(ctx context.Context, dst interface{}) (err error)
		// 偏移 排序
		Order(order string) SQLCommand
		Offset(offset int) SQLCommand
		Limit(limit int) SQLCommand

		// 数量
		Count(ctx context.Context) (count int, err error)
		// 批量插入
		BatchInsert(ctx context.Context, table string, rowsIn interface{}, rowsMethod BatchInsertFunc, columns []string, option ...BatchInsertOption) (err error)
	}

	BatchInsertFunc func(i int, insert func(...interface{}))

	PageParam struct {
		No    int
		Count int
		Order string
		// 可能有场景需要 但是可以通过改造select自动识别
		Distinct string
	}

	WRSQLConfig struct {
		Write SQLConfig   `json:"write"`
		Reads []SQLConfig `json:"reads"`
	}

	SQLConfig struct {
		Host              string `json:"host"`
		Port              int    `json:"port"`
		DB                string `json:"db"`
		User              string `json:"user"`
		Password          string `json:"password"`
		Location          string `json:"location"`
		InterpolateParams bool   `json:"interpolate_params"` // 允许使用进程内占位符替换 能够避免自动切换stmt模式 提升IO 在某些字符集场景有sql注入风险

		ExtraParams string `json:"extra"` // 额外参数 key=value&key=value模式

		MaxIdleConns  int `json:"max_idle_conns"`   // 最大空闲连接数
		MaxOpenConns  int `json:"max_open_conn"`    // 最大连接数
		MaxLifeTimeMs int `json:"max_life_time_ms"` // 最大连接生命时间 毫秒
		MaxIdleTimeMs int `json:"max_idle_time_ms"` // 最大连接空闲时间 毫秒

		DisableInterceptor bool `json:"disable_interceptor"`  // 禁用埋点 将会丧失错误日志、监控、链路追踪等功能
		SlowLogThresholdMs int  `json:"slowlog_threshold_ms"` // 慢日志阀值，超过该阀值的请求，将被记录到慢日志中，毫秒

		// 已废弃字段
		DeprecatedMaxLifeTimeMs      int `json:"max_lifetime,omitempty"`   // deprecated: 最大连接生命时间 毫秒 已废弃 使用 max_life_time_ms
		DeprecatedSlowLogThresholdMs int `json:"slow_threshold,omitempty"` // deprecated: 慢日志阀值，超过该阀值的请求，将被记录到慢日志中，毫秒 使用 slowlog_threshold_ms

	}

	MultiWRSQLConfig map[string]WRSQLConfig
	MultiSQLConfig   map[string]SQLConfig
)

const conUrl = "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s"

func (cfg *SQLConfig) Init() {
	procs := runtime.GOMAXPROCS(-1)

	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 10 + procs*10
	}

	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10 + procs*5
	}

	if cfg.MaxLifeTimeMs == 0 {
		cfg.MaxLifeTimeMs = cfg.DeprecatedMaxLifeTimeMs
	}

	if cfg.MaxLifeTimeMs <= 0 {
		cfg.MaxLifeTimeMs = 120000
	}

	if cfg.MaxIdleTimeMs <= 0 {
		cfg.MaxIdleTimeMs = 120000
	}

	if cfg.SlowLogThresholdMs == 0 {
		cfg.SlowLogThresholdMs = cfg.DeprecatedSlowLogThresholdMs
	}

	if cfg.SlowLogThresholdMs <= 0 {
		cfg.SlowLogThresholdMs = 1500
	}

	if len(cfg.Location) == 0 {
		cfg.Location = "Asia/Shanghai"
	}
}

func (cfg SQLConfig) ConnectionUrl() string {
	conStr := fmt.Sprintf(
		conUrl,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		url.QueryEscape(cfg.Location),
	)

	if cfg.InterpolateParams {
		conStr += "&interpolateParams=true"
	}

	if len(cfg.ExtraParams) > 0 {
		conStr += "&" + strings.TrimPrefix(cfg.ExtraParams, "&")
	}
	return conStr
}

func init() {
	dialer := knet.WrapTcpDialer(time.Second, time.Second*120, nil)
	mysql.RegisterDialContext("tcp", func(ctx context.Context, addr string) (net.Conn, error) {
		return dialer(ctx, "tcp", addr)
	})
}
