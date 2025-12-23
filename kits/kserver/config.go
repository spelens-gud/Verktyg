package kserver

import (
	"context"
	"os"
	"strconv"

	"github.com/spelens-gud/Verktyg/implements/promhttp"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/imetrics"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
	"github.com/spelens-gud/Verktyg/kits/kgo"
	"github.com/spelens-gud/Verktyg/kits/kserver/graceful"
)

type Config struct {
	Addr                     string                 `json:"addr"`
	ReadTimeoutSeconds       int                    `json:"read_timeout_seconds"`
	ReadHeaderTimeoutSeconds int                    `json:"read_header_timeout_seconds"`
	WriteTimeoutSeconds      int                    `json:"write_timeout_seconds"`
	IdleTimeoutSeconds       int                    `json:"idle_timeout_seconds"`
	CloseWaitSeconds         int                    `json:"close_wait_seconds"`
	GracefulRestart          bool                   `json:"graceful_restart"`
	EnableAsyncLogger        bool                   `json:"enable_async_logger"`
	Metrics                  imetrics.ServerMetrics `json:"-"`

	BaseContext context.Context `json:"-"`
}

func (config *Config) Init() {
	if len(config.Addr) == 0 {
		config.Addr = ":http"
	}

	if config.Metrics == nil {
		config.Metrics = promhttp.DefaultServerMetrics
	}

	if config.WriteTimeoutSeconds == 0 {
		config.WriteTimeoutSeconds = 120
	}

	if config.IdleTimeoutSeconds == 0 {
		config.IdleTimeoutSeconds = 120
	}

	if config.ReadTimeoutSeconds == 0 {
		config.ReadTimeoutSeconds = 120
	}

	if config.ReadHeaderTimeoutSeconds == 0 {
		config.ReadHeaderTimeoutSeconds = 120
	}

	if config.CloseWaitSeconds == 0 {
		config.CloseWaitSeconds, _ = strconv.Atoi(
			kgo.InitStrings(
				os.Getenv(iconfig.EnvKeyServerCloseWaitSeconds),
				os.Getenv(iconfig.EnvKeyServerCloseWaitSecondsOld),
			),
		)
		if config.CloseWaitSeconds == 0 {
			config.CloseWaitSeconds = defaultCloseWaitSeconds
		}
	}

	if !config.GracefulRestart {
		config.GracefulRestart = graceful.EnvGracefulReload() || len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0
	}

	if !config.EnableAsyncLogger {
		config.EnableAsyncLogger = iconfig.GetEnv().IsProduction() && envflag.IsEmpty(iconfig.EnvKeyLogDisableAsync)
	}
}
