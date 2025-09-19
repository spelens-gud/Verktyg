package bufferedwriter

import (
	"io"
	"testing"
)

func BenchmarkW(b *testing.B) {
	d := []byte{'x'}
	writer := NewBypassBufferedWriter(io.Discard, NewZapBufferedWriter)
	// nolint
	defer writer.Close()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = writer.Write(d)
		}
	})
}
