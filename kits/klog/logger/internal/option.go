package internal

import (
	"io"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
)

type (
	option struct {
		level          ilog.Level
		loggerProvider ilog.LoggerProvider
		writerWrappers []func(writer io.Writer) io.Writer
		output         io.Writer
		initPatch      []ilog.LoggerPatch
		contextPatch   []ilog.LoggerPatch
	}
)

func RegisterInitPatch(p ...ilog.LoggerPatch) {
	getRuntime().update(func(opt *option) {
		opt.initPatch = append(opt.initPatch, p...)
	})
}

func RegisterContextPatch(p ...ilog.LoggerPatch) {
	r := getRuntime()
	r.Lock()
	r.option.contextPatch = append(r.option.contextPatch, p...)
	r.Unlock()
}

func RegisterWriterWrapper(w ...func(writer io.Writer) io.Writer) {
	getRuntime().update(func(opt *option) {
		opt.writerWrappers = append(opt.writerWrappers, w...)
	})
}

func SetOutput(writer io.Writer) {
	getRuntime().update(func(opt *option) {
		opt.output = writer
	})
}

func SetProvider(p ilog.LoggerProvider) {
	getRuntime().update(func(opt *option) {
		opt.loggerProvider = p
	})
}

func SetLevel(level ilog.Level) {
	getRuntime().update(func(opt *option) {
		opt.level = level
	})
}
