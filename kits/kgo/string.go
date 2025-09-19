package kgo

import (
	"reflect"
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
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	// nolint
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
	}))
}
