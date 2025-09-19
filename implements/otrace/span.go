package otrace

import (
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	"git.bestfulfill.tech/devops/go-core/interfaces/itrace"
)

type Span struct {
	opentracing.Span
}

func (s *Span) SetPeer(peer string) {
	ext.PeerAddress.Set(s.Span, peer)
}

func (s *Span) SetLayer(i int) {
	// nolint
	s.Span.SetTag("layer", itrace.GetSpanLayerName(i))
}

func (s *Span) SetOperationName(name string) {
	s.Span.SetOperationName(name)
}

func (s *Span) SetComponent(i int) {
	s.Tag(itrace.TagComponentID, strconv.Itoa(i))
	s.Tag(string(ext.Component), itrace.GetComponentName(i))
}

func (s *Span) Tag(key, value string) {
	// nolint
	s.Span.SetTag(key, value)
}

func (s *Span) Log(level, msg string, fields map[string]interface{}) {
	// nolint
	s.Span.LogFields(parseLogFields(level, msg, fields)...)
}

func parseLogFields(level, msg string, fields map[string]interface{}) (logFields []log.Field) {
	if len(msg) > 0 {
		logFields = make([]log.Field, 0, len(fields)+2)
		logFields = append(logFields, log.Message(msg))
		if len(level) > 0 {
			logFields = append(logFields, log.String("level", level))
		}
	} else {
		logFields = make([]log.Field, 0, len(fields))
	}

	if len(fields) > 0 {
		tmp := make([]interface{}, 0, len(fields)*2)
		for k, v := range fields {
			tmp = append(tmp, k, v)
		}
		f, _ := log.InterleavedKVToFields(tmp...)
		logFields = append(logFields, f...)
	}
	return
}

func (s *Span) Error(err error, fields map[string]interface{}) {
	ext.LogError(s.Span, err, parseLogFields("", "", fields)...)
}
