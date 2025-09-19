package samplereport

import (
	"hash/crc32"

	"github.com/SkyAPM/go2sky"
)

const ratio = 100000

type Reporter struct {
	go2sky.Reporter
	sampleRate float64
	base       uint32
}

func NewReporter(r go2sky.Reporter, sampleRate float64) *Reporter {
	return &Reporter{sampleRate: sampleRate, Reporter: r, base: uint32(sampleRate * ratio)}
}

func (r Reporter) Send(spans []go2sky.ReportedSpan) {
	spanSize := len(spans)
	if spanSize < 1 {
		return
	}
	rootSpan := spans[spanSize-1]
	rootCtx := rootSpan.Context()
	count := crc32.ChecksumIEEE([]byte(rootCtx.TraceID)) % ratio
	if count < r.base {
		r.Reporter.Send(spans)
	}
}

var _ go2sky.Reporter = &Reporter{}
