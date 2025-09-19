package logger

import (
	"context"
	"io"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
	"git.bestfulfill.tech/devops/go-core/internal/incontext"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger/internal"
)

const (
	ctxKey = incontext.Key("logger")
)

func Writer() io.Writer {
	return internal.Writer()
}

// 设置日志级别 会重新初始化logger
func SetLevel(level ilog.Level) {
	internal.SetLevel(level)
}

// 设置日志输出writer 会重新初始化logger
func SetOutput(writer io.Writer) {
	internal.SetOutput(writer)
}

// 设置日志工厂方法 会重新初始化logger
func SetProvider(loggerFactory ilog.LoggerProvider) {
	internal.SetProvider(loggerFactory)
}

// 注册固有字段 会重新初始化logger
func RegisterConstPatch(p ...ilog.LoggerPatch) {
	internal.RegisterInitPatch(p...)
}

// 注册全局上下文字段 每次执行FromContext时 按照Patch链更新Logger
func RegisterGlobalContextPatch(p ...ilog.LoggerPatch) {
	internal.RegisterContextPatch(p...)
}

func RegisterWriterWrapper(w ...func(writer io.Writer) io.Writer) {
	internal.RegisterWriterWrapper(w...)
}

// 注册上下文字段 每次执行FromContext时 按照Patch链更新Logger
func WithContextPatch(ctx context.Context, p ...ilog.LoggerPatch) context.Context {
	p = append(internal.ExtPatchFromContext(ctx), p...)
	return internal.ExtPatchWithContext(ctx, p...)
}

func Flush() {
	internal.Flush()
}

func WithField(ctx context.Context, key string, val interface{}) context.Context {
	return WithContext(ctx, FromContext(ctx).WithField(key, val))
}

func AddCallerSkip(ctx context.Context, skip int) context.Context {
	return WithContext(ctx, FromContext(ctx).AddCallerSkip(skip))
}

func WithTags(ctx context.Context, tags ...string) context.Context {
	return WithContext(ctx, FromContext(ctx).WithTag(tags...))
}

func WithFields(ctx context.Context, fields map[string]interface{}) context.Context {
	return WithContext(ctx, FromContext(ctx).WithFields(fields))
}

func FromBackground() ilog.Logger {
	return FromContext(ctxBackground)
}

func WithContext(ctx context.Context, logger ilog.Logger) context.Context {
	return ctxKey.WithValue(ctx, logger)
}

func FromContext(ctx context.Context) ilog.Logger {
	logger, _ := ctxKey.Value(ctx).(ilog.Logger)
	if logger == nil {
		logger = internal.InitContextLogger(ctx)
	}
	return logger
}
