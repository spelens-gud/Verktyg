package bufferedwriter

import (
	"io/ioutil"
	"testing"
)

func BenchmarkW(b *testing.B) {
	d := []byte{'x'}
	writer := NewBypassBufferedWriter(ioutil.Discard, NewZapBufferedWriter)
	defer writer.Close()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = writer.Write(d)
		}
	})
}
