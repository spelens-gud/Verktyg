package skytrace

import (
	"github.com/SkyAPM/go2sky"
	jsoniter "github.com/json-iterator/go"
)

func NewLogReporter(lg Logger) go2sky.Reporter {
	return &logReporter{lg: lg}
}

type Logger interface {
	Infof(format string, args ...interface{})
}

type logReporter struct {
	lg Logger
}

func (lr *logReporter) Boot(service string, serviceInstance string) {}

func (lr *logReporter) Send(spans []go2sky.ReportedSpan) {
	b, err := jsoniter.ConfigFastest.Marshal(spans)
	if err != nil {
		lr.lg.Infof("error: %s", err)
		return
	}
	root := spans[len(spans)-1]
	lr.lg.Infof("segment-%v: %s \n", root.Context().SegmentID, b)
}

func (lr *logReporter) Close() {
	lr.lg.Infof("close log reporter")
}
