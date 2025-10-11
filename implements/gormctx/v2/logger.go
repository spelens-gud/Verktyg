package gormctx

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type gorm2Logger struct {
}

func newGorm2Logger() gormLogger.Interface {
	return &gorm2Logger{}
}

// LogMode log level
func (l *gorm2Logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

// Info print info
func (l *gorm2Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.FromContext(ctx).WithTag("MYSQL").Infof(msg, data...)
}

// Warn print warn messages
func (l *gorm2Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.FromContext(ctx).WithTag("MYSQL").Warnf(msg, data...)
}

// Error print error messages
func (l *gorm2Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.FromContext(ctx).WithTag("MYSQL").WithTag("MYSQL_ERR").Errorf(msg, data...)
}

// Trace print sql message
func (l *gorm2Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.FromContext(ctx).WithTag("MYSQL").Warnf("%s [%.3fms] [rows:%v] %s", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		} else {
			logger.FromContext(ctx).WithTag("MYSQL").WithTag("MYSQL_ERR").Errorf("%s [%.3fms] [rows:%v] %s", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	} else {
		logger.FromContext(ctx).WithTag("MYSQL").Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
