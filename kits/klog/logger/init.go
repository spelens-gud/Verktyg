package logger

import (
	"context"
	"io"
	"os"

	"go.uber.org/zap/zapcore"

	"github.com/spelens-gud/Verktyg/implements/zaplog"
	"github.com/spelens-gud/Verktyg/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg/interfaces/ilog"
	"github.com/spelens-gud/Verktyg/interfaces/itrace"
	"github.com/spelens-gud/Verktyg/kits/kenv/envflag"
	"github.com/spelens-gud/Verktyg/kits/klog/logger/internal"
)

var (
	defaultProvider ilog.LoggerProvider = zaplog.Provider{}
	ctxBackground                       = context.Background()
)

func init() {
	// 异步输出
	initWriterWrapper()
	// 初始化日志级别
	initLevel()
	// 输出文件配置
	initOutput()
	// 注册固定字段
	initAppNameField()
	// 设置默认日志库
	initDefaultProvider()
	// 实例化补丁
	initPatchers()
}

const FieldRequestID = "request_id"

func initPatchers() {
	internal.RegisterContextPatch(tracePatch{}, FieldPatch{FieldRequestID: func(ctx context.Context) interface{} {
		if ctx == ctxBackground {
			return nil
		}
		if ret := itrace.FromContext(ctx); len(ret) > 0 {
			return ret
		}
		return nil
	}})
}

func initOutput() {
	if fireDir := os.Getenv(iconfig.EnvKeyLogFileDir); len(fireDir) > 0 {
		if err := SetFileOutput(func(option *FileLogOption) {
			option.Dir = fireDir
		}); err != nil {
			panic(err)
		}
	} else {
		SetOutput(os.Stdout)
	}
}

func initDefaultProvider() {
	SetProvider(defaultProvider)
}

func initWriterWrapper() {
	// 对文件类型writer加锁避免潜在的并发读写fd造成阻塞问题
	if !envflag.IsEmpty(iconfig.EnvKeyLogDisableFileMutex) {
		internal.RegisterWriterWrapper(func(writer io.Writer) io.Writer {
			if _, isFile := writer.(interface {
				Stat() (os.FileInfo, error)
			}); isFile {
				writer = zapcore.Lock(zapcore.AddSync(writer))
			}
			return writer
		})
	}

}

func initAppNameField() {
	internal.RegisterInitPatch(FieldPatch{"application": func(ctx context.Context) interface{} {
		return iconfig.GetApplicationName()
	}})
}

func initLevel() {
	level := os.Getenv(iconfig.EnvKeyLogLevel)
	if len(level) == 0 {
		if iconfig.GetEnv().IsDevelopment() {
			level = "debug"
		} else {
			level = "info"
		}
	}
	internal.SetLevel(ilog.ParseLevel(level))
}
