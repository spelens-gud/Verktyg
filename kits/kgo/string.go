package kgo

import (
	"unsafe"
)

func InitStringPtr(in *string, backs ...string) {
	out := *in
	if len(out) > 0 {
		return
	}
	*in = InitStrings(backs...)
}

func InitStrings(strings ...string) string {
	for _, str := range strings {
		if len(str) > 0 {
			return str
		}
	}
	return ""
}

func UnsafeBytes2string(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
