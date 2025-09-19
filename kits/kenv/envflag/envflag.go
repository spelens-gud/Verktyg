package envflag

import (
	"os"
	"sync"
)

func BoolOnceFrom(key string) func() bool {
	var (
		o sync.Once
		b *bool
	)
	return func() bool {
		o.Do(func() {
			tmp := IsNotEmpty(key)
			b = &tmp
		})
		return *b
	}
}

func BoolFrom(key string) func() bool {
	return func() bool {
		return IsNotEmpty(key)
	}
}

func IsNotEmpty(key string) bool {
	return !IsEmpty(key)
}

func IsEmpty(key string) bool {
	return len(os.Getenv(key)) == 0
}
