package internal

import (
	"runtime"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
)

func init() {
	if _, f, _, ok := runtime.Caller(0); ok {
		ilog.RegisterWrapLogger(f)
	}
}

type wrappedLogger struct {
	ilog.Logger
	OnLog func(level ilog.Level, msg string, args ...interface{}) (doLog bool)
}

func WrapLogger(logger ilog.Logger, OnLog func(level ilog.Level, msg string, args ...interface{}) bool) ilog.Logger {
	return wrapLogger(logger.AddCallerSkip(1), OnLog)
}

func wrapLogger(logger ilog.Logger, OnLog func(level ilog.Level, msg string, args ...interface{}) bool) ilog.Logger {
	return wrappedLogger{Logger: logger, OnLog: OnLog}
}

func (w wrappedLogger) WithTag(tags ...string) ilog.Logger {
	return wrappedLogger{
		Logger: w.Logger.WithTag(tags...),
		OnLog:  w.OnLog,
	}
}

func (w wrappedLogger) WithField(key string, val interface{}) ilog.Logger {
	return wrappedLogger{
		Logger: w.Logger.WithField(key, val),
		OnLog:  w.OnLog,
	}
}

func (w wrappedLogger) WithFields(fields map[string]interface{}) ilog.Logger {
	return wrappedLogger{
		Logger: w.Logger.WithFields(fields),
		OnLog:  w.OnLog,
	}
}

func (w wrappedLogger) AddCallerSkip(skip int) ilog.Logger {
	return wrappedLogger{
		Logger: w.Logger.AddCallerSkip(skip),
		OnLog:  w.OnLog,
	}
}

func (w wrappedLogger) Debug(msg string) {
	if w.OnLog(ilog.Debug, msg) {
		w.Logger.Debug(msg)
	}
}

func (w wrappedLogger) Debugf(msg string, args ...interface{}) {
	if w.OnLog(ilog.Debug, msg, args...) {
		w.Logger.Debugf(msg, args...)
	}
}

func (w wrappedLogger) Info(msg string) {
	if w.OnLog(ilog.Info, msg) {
		w.Logger.Info(msg)
	}
}

func (w wrappedLogger) Infof(msg string, args ...interface{}) {
	if w.OnLog(ilog.Info, msg, args...) {
		w.Logger.Infof(msg, args...)
	}
}

func (w wrappedLogger) Warn(msg string) {
	if w.OnLog(ilog.Warn, msg) {
		w.Logger.Warn(msg)
	}
}

func (w wrappedLogger) Warnf(msg string, args ...interface{}) {
	if w.OnLog(ilog.Warn, msg, args...) {
		w.Logger.Warnf(msg, args...)
	}
}

func (w wrappedLogger) Error(msg string) {
	if w.OnLog(ilog.Error, msg) {
		w.Logger.Error(msg)
	}
}

func (w wrappedLogger) Errorf(msg string, args ...interface{}) {
	if w.OnLog(ilog.Error, msg, args...) {
		w.Logger.Errorf(msg, args...)
	}
}
