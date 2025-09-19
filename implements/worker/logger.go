package worker

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
)

const logTagCronWorker = "CRON_WORKER"

type adapterLogger struct {
	lg ilog.Logger
}

func (a adapterLogger) Info(msg string, keysAndValues ...interface{}) {
	a.lg.WithFields(a.initFields(keysAndValues)).Infof(msg)
}

func (a adapterLogger) initFields(keysAndValues []interface{}) map[string]interface{} {
	fields := make(map[string]interface{}, (len(keysAndValues)+1)/2)
	l := len(keysAndValues)
	for i := 0; i < l; i += 2 {
		if i > l-1 {
			break
		}
		fields[cast.ToString(keysAndValues[i])] = keysAndValues[i+1]
	}
	return fields
}

func (a adapterLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	a.lg.WithFields(a.initFields(keysAndValues)).Errorf(msg+" error: %s", err.Error())
}

func WithLogger(lg ilog.Logger) cron.Option {
	return cron.WithLogger(adapterLogger{lg: lg.WithTag(logTagCronWorker).AddCallerSkip(1)})
}
