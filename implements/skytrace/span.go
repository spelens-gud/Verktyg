package skytrace

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/SkyAPM/go2sky"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
)

type Span struct {
	go2sky.Span
	end bool
}

func (s *Span) Tag(key, value string) {
	if s.Span == nil {
		return
	}
	s.Span.Tag(go2sky.Tag(key), value)
}

func (s *Span) Log(level, msg string, fields map[string]interface{}) {
	s.Span.Log(time.Now(), parseLogFields("message", level, msg, fields)...)
}

func parseLogFields(msgKey, level, msg string, fields map[string]interface{}) []string {
	kv := make([]string, 0, (len(fields)+2)*2)
	kv = append(kv, msgKey, msg)
	if len(level) > 0 {
		kv = append(kv, "level", level)
	}
	for k, v := range fields {
		kv = append(kv, k, fmt.Sprint(v))
	}
	return kv
}

func (s *Span) Error(err error, fields map[string]interface{}) {
	s.Span.Error(time.Now(), parseLogFields("error", "error", err.Error(), fields)...)
}

func (s *Span) Finish() {
	if s.Span == nil {
		return
	}
	if s.end {
		return
	}
	s.end = true
	// nolint
	s.Span.End()
}

func (s *Span) SetLayer(i int) {
	// nolint
	s.Span.SetSpanLayer(v3.SpanLayer(i))
}

func (s *Span) SetPeer(peer string) {
	s.Span.SetPeer(peer)
	s.Tag(string(ext.PeerAddress), peer)
}

func (s *Span) SetComponent(i int) {
	s.Span.SetComponent(int32(i))
	s.Tag(itrace.TagComponentID, strconv.Itoa(i))
	s.Tag(string(ext.Component), itrace.GetComponentName(i))
}

const fieldName = "TraceID"

func setSpanTraceID(span go2sky.Span, requestID string) {
	v := reflect.ValueOf(span).Elem()
	if _, ok := v.Type().FieldByName(fieldName); ok {
		v.FieldByName(fieldName).Set(reflect.ValueOf(requestID))
	}
	if rSpan, ok := span.(go2sky.ReportedSpan); ok {
		for _, ref := range rSpan.Refs() {
			ref.TraceID = requestID
		}
	}
}

func getSpanTraceID(span go2sky.Span) string {
	s, ok := span.(go2sky.ReportedSpan)
	if ok {
		return s.Context().TraceID
	}
	return ""
}
