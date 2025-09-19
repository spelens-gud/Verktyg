package main

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	language_agent "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

func TestClient(t *testing.T) {
	for range time.NewTicker(time.Second).C {
		str, _ := jsoniter.MarshalToString(&language_agent.SegmentObject{
			TraceId:        uuid.New().String(),
			TraceSegmentId: uuid.New().String(),
			Spans: []*language_agent.SpanObject{
				{
					SpanId:               0,
					ParentSpanId:         -1,
					StartTime:            1606191965487,
					EndTime:              1606191965499,
					Refs:                 nil,
					OperationName:        uuid.New().String(),
					Peer:                 uuid.New().String(),
					SpanType:             2,
					SpanLayer:            2,
					ComponentId:          5,
					IsError:              false,
					Tags:                 nil,
					Logs:                 nil,
					SkipAnalysis:         false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
				{
					SpanId:               1,
					ParentSpanId:         0,
					StartTime:            1606191965499,
					EndTime:              1606191965599,
					Refs:                 nil,
					OperationName:        "zxbcbxc",
					Peer:                 "zbcbc",
					SpanType:             2,
					SpanLayer:            2,
					ComponentId:          5,
					IsError:              false,
					Tags:                 nil,
					Logs:                 nil,
					SkipAnalysis:         false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			Service:         "xxxx",
			ServiceInstance: "vxcvzcxvcx",
			IsSizeLimited:   false,
		})
		resp, err := http.Post("http://localhost:20202", "application/json", bytes.NewBufferString(str))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(resp.StatusCode)
	}
}
