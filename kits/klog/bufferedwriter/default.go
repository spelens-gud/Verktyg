package bufferedwriter

import (
	"io"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap/zapcore"
)

type zapBuffWriter struct {
	zapcore.BufferedWriteSyncer
}

func (w *zapBuffWriter) Flush() {
	_ = w.BufferedWriteSyncer.Sync()
}

type nonFlusher struct {
	io.Writer
}

func (n nonFlusher) Flush() {}

func NewZapBufferedWriter(writer io.Writer) BufferedWriter {
	if _, wrapped := writer.(*zapBuffWriter); wrapped {
		return nonFlusher{writer}
	}

	return &zapBuffWriter{BufferedWriteSyncer: zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(writer),
		FlushInterval: time.Second * 5,
	}}
}

type BufferedWriter interface {
	io.Writer
	Flush()
}

type bypassBufferedWriter struct {
	patched BufferedWriter
	origin  io.Writer
	closed  atomic.Bool
}

func (b *bypassBufferedWriter) Write(p []byte) (n int, err error) {
	if b.closed.Load() {
		return b.origin.Write(p)
	}
	return b.patched.Write(p)
}

func (b *bypassBufferedWriter) Close() error {
	if !b.closed.Swap(true) {
		b.patched.Flush()
	}
	return nil
}

func NewBypassBufferedWriter(writer io.Writer, patch ...func(writer io.Writer) BufferedWriter) io.WriteCloser {
	if len(patch) == 0 {
		panic("invalid patch")
	}
	patched := writer
	for _, p := range patch {
		patched = p(patched)
	}
	return &bypassBufferedWriter{
		patched: patched.(BufferedWriter),
		origin:  writer,
	}
}
