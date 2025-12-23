package kvcfgloader

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/spelens-gud/Verktyg/implements/cfgloader/sdk_acm"
)

func init() {
	go func() {
		panic(http.ListenAndServe(":8082", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(`{}`))
			//writer.WriteHeader(404)

		})))
	}()
}

var loader2 = NewCacheConfigLoader(
	NewAcmKVConfig("xx", sdk_acm.NewClient("http://localhost:8082")),
	WithDefaultExpiration(time.Second*5),
	WithBlocksNum(8),
	DisableMustGet(),
	LazyUpdate(),
	WithSrcTimeOut(time.Second*2),
)

func BenchmarkStore(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = loader2.GetWithContext(ctx, fmt.Sprintf("%v.json", rand.Intn(1000000)))
		}
	})
	_ = ctx
}

func TestStore(t *testing.T) {
	loader := NewCacheConfigLoader(
		NewAcmKVConfig("xx", sdk_acm.NewClient("http://localhost:8082")),
		LazyUpdate(),
		WithDefaultExpiration(time.Second*5),
	)
	ctx := context.Background()
	for {
		ret, err := loader.GetWithContext(ctx, "xxx###xxxx")
		if err != nil {
			t.Log(err)
		} else {
			t.Log(ret)
		}
		time.Sleep(time.Second)
	}
}
