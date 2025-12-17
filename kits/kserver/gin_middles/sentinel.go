package gin_middles

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	sentinelgin "github.com/alibaba/sentinel-golang/adapter/gin"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/system"
	sentinellog "github.com/alibaba/sentinel-golang/logging"
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/kits/kenv/envflag"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

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

// TODO: 分离初始化和中间件逻辑
func (opt *SentinelOption) loadCpuOverloadRule() (err error) {
	// 初始化
	if err = sentinel.InitDefault(); err != nil {
		err = fmt.Errorf("初始化sentinel失败: %v", err)
		return
	}

	sentinellog.ResetDefaultLogger(logger.NewStandardLogger(), "SENTINEL")

	rule := &system.SystemRule{
		ID:           1,
		MetricType:   system.CpuUsage,
		Strategy:     system.BBR,
		TriggerCount: opt.CpuUsageTriggerCount,
	}

	// 负载策略
	if opt.DisableBBR {
		rule.Strategy = system.NoAdaptive
	}

	// 加载策略
	_, err = system.LoadRules([]*system.SystemRule{rule})
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
