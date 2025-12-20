package kgrpc

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/kits/kgo"
)

type Config struct {
	Addr              string          `json:"addr,omitempty"`               // 监听地址
	MaxRecvMsgSize    int             `json:"max_recv_msg_size,omitempty"`  // 接收消息最大长度
	MaxSendMsgSize    int             `json:"max_send_msg_size,omitempty"`  // 发送消息最大长度
	ConnectionTimeout int             `json:"connection_timeout,omitempty"` // 连接超时
	CloseWaitSeconds  int             `json:"close_wait_seconds,omitempty"` // 关闭等待时间
	EnableReflection  bool            `json:"enable_reflection,omitempty"`  // 是否启用反射
	BaseContext       context.Context `json:"base_context,omitempty"`       // 基础上下文
}

const (
	// 默认最大接收消息长度
	defaultMaxMsgSize = 4 * 1024 * 1024 // 4MB
	// 默认连接超时
	defaultConnectionTimeout = 120
	// 默认关闭等待时间
	defaultCloseWaitSeconds = 10
)

// Init method    初始化配置，填充默认值.
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
