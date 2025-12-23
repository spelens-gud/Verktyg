package internal

import (
	"io"

	"github.com/spelens-gud/Verktyg/kits/kgo"
)

type stdLoggerWriter struct{}

func StandardLogWriter() io.Writer { return stdLoggerWriter{} }

func (stdLoggerWriter) Write(p []byte) (n int, err error) {
	lp := len(p)
	if lp > 0 && p[lp-1] == '\n' {
		p = p[:lp-1]
	}
	getRuntime().getInstance().GetStdLogger().Info(kgo.UnsafeBytes2string(p))
	return lp, nil
}
