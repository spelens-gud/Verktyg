package buffpool

import (
	"bytes"
	"testing"
)

var tmp = []byte("Sgasga")

func BenchmarkPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bf := GetBytesBuffer()
			if bf.Len() != 0 {
				b.Fatal("dirty buff")
			}
			bf.Write(tmp)
			PutBytesBuffer(bf)
		}
	})
}

func BenchmarkRaw(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bf := bytes.NewBuffer(nil)
			if bf.Len() != 0 {
				b.Fatal("dirty buff")
			}
			bf.Write(tmp)
		}
	})
}
