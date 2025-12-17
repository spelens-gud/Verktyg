package internal

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/spelens-gud/Verktyg/interfaces/ilog"
)

type instance struct {
	ilog.Logger
	sync.Once
	option *option

	Writer        io.Writer
	ContextLogger ilog.Logger
	StdLogger     ilog.Logger
	Flusher       []Flusher
}

type Flusher interface {
	Flush()
}

func (i *instance) Flush() {
	for _, flusher := range i.Flusher {
		flusher.Flush()
		if closer, ok := flusher.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}

func (i *instance) GetLogger() ilog.Logger {
	return i.initLogger().Logger
}

func (i *instance) GetWriter() io.Writer {
	return i.initLogger().Writer
}

func (i *instance) GetStdLogger() ilog.Logger {
	return i.initLogger().StdLogger
}

func (i *instance) GetCtxBackgroundLogger() ilog.Logger {
	return i.initLogger().ContextLogger
}

func (i *instance) initLogger() *instance {
	i.Do(func() {
		opt := i.option
		if opt.loggerProvider == nil || opt.output == nil {
			panic("logger provider unset")
		}

		i.Writer = opt.output

		for _, w := range opt.writerWrappers {
			i.Writer = w(i.Writer)
			if flusher, ok := i.Writer.(Flusher); ok {
				i.Flusher = append(i.Flusher, flusher)
			}
		}

		logger := opt.loggerProvider.Init(i.Writer)

		for _, p := range opt.initPatch {
			logger = p.Patch(context.Background(), logger)
		}

		logger.WithFields(map[string]interface{}{
			"provider":  fmt.Sprintf("%T", opt.loggerProvider),
			"logger":    fmt.Sprintf("%T", logger),
			"log_level": opt.level.String(),
		}).Info("logger init")

		if i.Logger = logger; !opt.level.EnableAll() {
			// nolint
			i.Logger = i.Logger.AddCallerSkip(1)
		}
		i.ContextLogger = initContextLogger(ctxBackground, i.Logger)
		i.StdLogger = logger.AddCallerSkip(3)
	})
	return i
}
