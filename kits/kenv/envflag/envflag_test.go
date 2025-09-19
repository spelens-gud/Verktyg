package envflag

import (
	"testing"
)

func BenchmarkEnv(b *testing.B) {
	env := BoolOnceFrom("test")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = env()
		}
	})
}

func BenchmarkEnv2(b *testing.B) {
	env := BoolFrom("test")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = env()
		}
	})
}

func BenchmarkEnv3(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = IsEmpty("test")
		}
	})
}
