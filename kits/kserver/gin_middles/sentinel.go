package gin_middles

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/system"
	sentinellog "github.com/alibaba/sentinel-golang/logging"
	sentinelgin "github.com/alibaba/sentinel-golang/pkg/adapters/gin"
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

// sentinelLogger struct 适配 sentinel Logger 接口
type sentinelLogger struct {
	stdLogger *log.Logger
}

// Debug method 调试日志
func (l *sentinelLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.stdLogger.Printf("[DEBUG] "+msg, keysAndValues...)
}

// DebugEnabled method 是否启用调试日志
func (l *sentinelLogger) DebugEnabled() bool {
	return false
}

// Info method 信息日志
func (l *sentinelLogger) Info(msg string, keysAndValues ...interface{}) {
	l.stdLogger.Printf("[INFO] "+msg, keysAndValues...)
}

// InfoEnabled method 是否启用信息日志
func (l *sentinelLogger) InfoEnabled() bool {
	return true
}

// Warn method 警告日志
func (l *sentinelLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.stdLogger.Printf("[WARN] "+msg, keysAndValues...)
}

// WarnEnabled method 是否启用警告日志
func (l *sentinelLogger) WarnEnabled() bool {
	return true
}

// Error method 错误日志
func (l *sentinelLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.stdLogger.Printf("[ERROR] "+msg+" error=%v", append(keysAndValues, err)...)
}

// ErrorEnabled method 是否启用错误日志
func (l *sentinelLogger) ErrorEnabled() bool {
	return true
}

var (
	once       sync.Once
	disableBBR = envflag.BoolFrom(iconfig.EnvKeyServerSentinelDisableBBR)
)

const defaultTriggerCount = 0.85

type SentinelOption struct {
	DisableBBR           bool
	CpuUsageTriggerCount float64
	FallbackStatus       int
}

// loadCpuOverloadRule method 加载 CPU 过载保护规则
func (opt *SentinelOption) loadCpuOverloadRule() (err error) {
	// 初始化
	if err = sentinel.InitDefault(); err != nil {
		err = fmt.Errorf("初始化sentinel失败: %v", err)
		return
	}

	// 设置自定义日志
	_ = sentinellog.ResetGlobalLogger(&sentinelLogger{stdLogger: logger.NewStandardLogger()})

	rule := &system.Rule{
		ID:           "1",
		MetricType:   system.CpuUsage,
		Strategy:     system.BBR,
		TriggerCount: opt.CpuUsageTriggerCount,
	}

	// 负载策略
	if opt.DisableBBR {
		rule.Strategy = system.NoAdaptive
	}

	// 加载策略
	_, err = system.LoadRules([]*system.Rule{rule})
	return
}

func Sentinel(opts ...func(*SentinelOption)) gin.HandlerFunc {
	triggerCount, err := strconv.ParseFloat(os.Getenv(iconfig.EnvKeyServerSentinelCpuTrigger), 64)
	if err != nil || triggerCount >= 1 || triggerCount <= 0 {
		triggerCount = defaultTriggerCount
	}

	opt := &SentinelOption{
		DisableBBR:           disableBBR(),
		CpuUsageTriggerCount: triggerCount,
		FallbackStatus:       http.StatusTooManyRequests,
	}

	for _, o := range opts {
		o(opt)
	}

	once.Do(func() {
		if err := opt.loadCpuOverloadRule(); err != nil {
			panic(err)
		}
	})

	return sentinelgin.SentinelMiddleware(sentinelgin.WithBlockFallback(func(c *gin.Context) {
		c.AbortWithStatus(opt.FallbackStatus)
	}))
}
