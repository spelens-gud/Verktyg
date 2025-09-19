package version

import (
	"runtime"
	"strings"
	"sync"
)

var o sync.Once

func getRuntimeVersion() string {
	if _, f, _, ok := runtime.Caller(1); ok {
		if tmp := strings.Split(f, "sy-core@"); len(tmp) > 1 {
			return strings.Split(tmp[1], "/")[0]
		}
	}
	return ""
}

func GetVersion() string {
	o.Do(func() {
		if rv := getRuntimeVersion(); len(rv) > 0 {
			version = rv
		}
	})
	return version
}
