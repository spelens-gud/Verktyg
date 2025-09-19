package log

import (
	"testing"
)

var tmp = []byte("Sgasga")

func BenchmarkPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bf := GetServerLoggerInfo()
			bf.Duration = string(tmp)
			doSomething(bf)
			PutServerLoggerInfo(bf)
		}
	})
}

func doSomething(stringer interface {
	Marshal() []byte
}) {
	_ = stringer.Marshal()
}

func BenchmarkRaw(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bf := &ServerLoggerInfo{}
			bf.Duration = string(tmp)
			doSomething(bf)
		}
	})
}
