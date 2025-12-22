// Package kgrpc 提供 gRPC 服务器的配置和启动支持.
// 包含服务器配置、优雅关闭、拦截器链等功能.
package kgrpc

import (
	"context"
	"os"
	"strconv"
	"time"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/kgo"
)

// Config struct gRPC 服务器配置.
type Config struct {
	// Addr 监听地址，格式为 :port 或 host:port.
	Addr string `json:"addr"`
	// MaxRecvMsgSize 最大接收消息大小（字节），默认 4MB.
	MaxRecvMsgSize int `json:"max_recv_msg_size"`
	// MaxSendMsgSize 最大发送消息大小（字节），默认 4MB.
	MaxSendMsgSize int `json:"max_send_msg_size"`
	// ConnectionTimeout 连接超时时间（秒），默认 120.
	ConnectionTimeout int `json:"connection_timeout"`
	// CloseWaitSeconds 优雅关闭等待时间（秒），默认 10.
	CloseWaitSeconds int `json:"close_wait_seconds"`
	// EnableReflection 是否启用 gRPC 反射服务，开发环境默认启用.
	EnableReflection bool `json:"enable_reflection"`
	// BaseContext 基础上下文.
	BaseContext context.Context `json:"-"`
}

const (
	defaultMaxMsgSize        = 4 * 1024 * 1024 // 4MB
	defaultConnectionTimeout = 120
	defaultCloseWaitSeconds  = 10
)

// Init method 初始化配置，填充默认值.
func (c *Config) Init() {
	if len(c.Addr) == 0 {
		c.Addr = ":50051"
	}

	if c.MaxRecvMsgSize == 0 {
		c.MaxRecvMsgSize = defaultMaxMsgSize
	}

	if c.MaxSendMsgSize == 0 {
		c.MaxSendMsgSize = defaultMaxMsgSize
	}

	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = defaultConnectionTimeout
	}

	if c.CloseWaitSeconds == 0 {
		c.CloseWaitSeconds, _ = strconv.Atoi(
			kgo.InitStrings(
				os.Getenv(iconfig.EnvKeyServerCloseWaitSeconds),
				os.Getenv(iconfig.EnvKeyServerCloseWaitSecondsOld),
			),
		)
		if c.CloseWaitSeconds == 0 {
			c.CloseWaitSeconds = defaultCloseWaitSeconds
		}
	}

	// 开发环境默认启用反射
	if !c.EnableReflection {
		c.EnableReflection = iconfig.GetEnv().IsDevelopment()
	}
}

// GetCloseWaitDuration method 获取优雅关闭等待时间.
func (c *Config) GetCloseWaitDuration() time.Duration {
	return time.Duration(c.CloseWaitSeconds) * time.Second
}

// GetConnectionTimeout method 获取连接超时时间.
func (c *Config) GetConnectionTimeout() time.Duration {
	return time.Duration(c.ConnectionTimeout) * time.Second
}
