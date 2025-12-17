package main

import (
	"context"
	"math/rand"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

type patch string

func (p patch) Patch(ctx context.Context, logger ilog.Logger) ilog.Logger {
	// 初始化logger时自动加上TAG
	return logger.WithTag(string(p))
}

type contextKey string

const (
	contextKeyClientIP = contextKey("client_ip")
	contextKeyUserID   = contextKey("user_id")
)

func init() {
	// 注册 自定义Patch
	// RegisterConstPatch 用于 固定logger变更 Patch方法中的context是context.Background 无法提取相关信息
	logger.RegisterConstPatch(patch("MY_TAG"))

	// 注册 通过context键值对 自动提取 添加的字段
	// RegisterGlobalContextPatch 用于 上下文logger变更 Patch方法中的context是FromContext时传入的ctx 可以提取上下文信息
	logger.RegisterGlobalContextPatch(logger.NewContextFieldPatch(map[string]interface{}{
		"ip":  contextKeyClientIP,
		"uid": contextKeyUserID,
	}))
}

func main() {
	ctx := context.Background()

	// 所有的日志输出都会带上patch里的变更
	logger.FromContext(ctx).Infof("test")
	// {"level":"info","msg":"test","tag":"MY_TAG"}

	// 在context内添加键值对 patch能自动提取对应的 键值并映射到日志字段中
	clientIPContext := context.WithValue(ctx, contextKeyClientIP, "127.0.0.1")
	logger.FromContext(clientIPContext).Infof("test context key field")
	// {"level":"info","msg":"test context key field","tag":"MY_TAG","ip":"127.0.0.1"}

	// 包含多个键值对
	multiContext := context.WithValue(clientIPContext, contextKeyUserID, rand.Intn(1000))
	logger.FromContext(multiContext).Infof("test multi context key field")
	// {"level":"info","msg":"test multi context key field","tag":"MY_TAG","ip":"127.0.0.1","uid":81}

}
