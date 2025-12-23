package redisx

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/redis/go-redis/v9"
	"go.uber.org/multierr"

	"github.com/spelens-gud/Verktyg/implements/promdb"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/itrace"
	"github.com/spelens-gud/Verktyg/internal/incontext"
	"github.com/spelens-gud/Verktyg/kits/kdb"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
	"github.com/spelens-gud/Verktyg/kits/ktrace/tracer"
)

var _ redis.Hook = &interceptor{}

type (
	interceptor struct {
		addr         string
		instanceName string
		errorLog     bool
	}

	ctxStore struct {
		cmd       string
		span      itrace.Span
		startTime time.Time
	}
)

// DialHook method 实现 redis.Hook 接口的 DialHook 方法
func (t *interceptor) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

// ProcessHook method 实现 redis.Hook 接口的 ProcessHook 方法
func (t *interceptor) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// Before 逻辑
		ctx, err := t.BeforeProcess(ctx, cmd)
		if err != nil {
			return err
		}

		// 执行实际命令
		err = next(ctx, cmd)

		// After 逻辑
		_ = t.AfterProcess(ctx, cmd)

		return err
	}
}

// ProcessPipelineHook method 实现 redis.Hook 接口的 ProcessPipelineHook 方法
func (t *interceptor) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		// Before 逻辑
		ctx, err := t.BeforeProcessPipeline(ctx, cmds)
		if err != nil {
			return err
		}

		// 执行实际命令
		err = next(ctx, cmds)

		// After 逻辑
		_ = t.AfterProcessPipeline(ctx, cmds)

		return err
	}
}

const (
	contextKeyGoRedis incontext.Key = "go-redis.hook"
)

type client interface {
	promdb.RedisStatsGetter
	redis.UniversalClient
}

func RegisterHookMetrics(cli client, traceAddr string, metricsName string, db int) {
	registerHookMetrics(cli, traceAddr, metricsName, db)
}

func registerHookMetrics(cli client, traceAddr string, metricsName string, db int) {
	// 加入链路追踪 监控 日志等
	hook := &interceptor{
		addr:     traceAddr,
		errorLog: envflag.IsEmpty(iconfig.EnvKeyLogDisableRedisError),
	}

	if db >= 0 {
		hook.instanceName = strconv.Itoa(db)
	} else {
		hook.instanceName = "cluster"
	}

	cli.AddHook(hook)
	// 监测
	promdb.RegisterRedisStatsMetrics(metricsName, cli)
}

func (t *interceptor) Log(ctx context.Context, s ctxStore, instance string, err error) {
	ctx = tracer.SpanWithContext(ctx, s.span)
	du := time.Since(s.startTime)

	lg := logger.FromContext(ctx).WithTag("REDIS").WithFields(map[string]interface{}{
		"cmd":      s.cmd,
		"instance": instance,
		"duration": du.String(),
		"cost":     du.Seconds(),
	})

	if err != context.Canceled {
		lg = lg.WithTag("REDIS_ERR")
		s.span.Error(err, lg.FieldData())
	}

	if t.errorLog {
		lg.Error(err.Error())
	}
}

func (t *interceptor) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	select {
	case <-ctx.Done():
		return ctx, ctx.Err()
	default:
	}

	return contextKeyGoRedis.WithValue(ctx, ctxStore{
		cmd:       cmd.String(),
		span:      t.startTraceSpan(ctx, cmd.Name(), cmd.String()),
		startTime: time.Now(),
	}), nil
}

func (t *interceptor) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	store, ok := contextKeyGoRedis.Value(ctx).(ctxStore)
	if !ok {
		return nil
	}

	defer store.span.Finish()

	err := cmd.Err()

	var (
		cmdName  = cmd.Name()
		duration = time.Since(store.startTime)
	)

	kdb.RedisMetrics(cmdName, t.addr, "", err, duration)

	if err != nil && err != redis.Nil {
		t.Log(ctx, store, t.addr, err)
	}

	// 队列监控
	if args := cmd.Args(); len(args) > 1 {
		switch strings.ToLower(cmdName) {
		case "lpush", "lpushx", "rpush", "rpushx":
			if topic, ok := args[1].(string); ok {
				kdb.RedisListProduceMetrics(t.addr, topic, err, duration)
			}
		case "lpop", "rpop", "blpop", "brpop":
			if topic, ok := args[1].(string); ok {
				kdb.RedisListConsumeMetrics(t.addr, topic, err, duration)
			}
		}
	}
	return nil
}

func (t *interceptor) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	select {
	case <-ctx.Done():
		return ctx, ctx.Err()
	default:
	}

	cmdStrs := make([]string, 0, len(cmds))
	statements := make([]string, 0, len(cmds))

	for _, cmd := range cmds {
		cmdStrs = append(cmdStrs, cmd.Name())
	}

	for _, cmd := range cmds {
		statements = append(statements, cmd.String())
	}

	cmd := "pipeline:" + strings.Join(cmdStrs, ",")

	return contextKeyGoRedis.WithValue(ctx, ctxStore{
		cmd:       cmd,
		span:      t.startTraceSpan(ctx, cmd, strings.Join(statements, ",")),
		startTime: time.Now(),
	}), nil
}

var errPipelineError = errors.New("pipeline error")

func (t *interceptor) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	store, ok := contextKeyGoRedis.Value(ctx).(ctxStore)
	if !ok {
		return nil
	}

	defer store.span.Finish()

	var errs []error

	for _, cmd := range cmds {
		if e := cmd.Err(); e != nil && e != redis.Nil {
			errs = append(errs, e)
		}
	}

	err := multierr.Combine(errs...)
	if err != nil {
		kdb.RedisMetrics("pipeline", t.addr, "", errPipelineError, time.Since(store.startTime))
		t.Log(ctx, store, t.addr, err)
	} else {
		kdb.RedisMetrics("pipeline", t.addr, "", nil, time.Since(store.startTime))
	}
	return nil
}

func (t *interceptor) startTraceSpan(ctx context.Context, name string, statement string) (span itrace.Span) {
	span = tracer.StartExitSpan(ctx, "REDIS_CLIENT:"+name, t.addr)

	span.SetComponent(itrace.ComponentRedis)
	span.SetLayer(itrace.SpanLayerCache)

	span.Tag(string(ext.DBStatement), statement)
	span.Tag(string(ext.DBType), "redis")
	span.Tag(string(ext.DBInstance), t.instanceName)
	span.Tag(itrace.TagDBStatementType, name)
	return
}
