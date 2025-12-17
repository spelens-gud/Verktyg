package cachew_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spelens-gud/Verktyg.git/implements/redisx"
	"github.com/spelens-gud/Verktyg.git/interfaces/iredis"
	"github.com/spelens-gud/Verktyg.git/kits/kgo/cachew"
	"github.com/spelens-gud/Verktyg.git/kits/kgo/cachew/redistorew"
)

type ret struct {
	Ret string `json:"ret"`
}

func startServer(ctx context.Context, addr string) {
	panic((&http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			time.Sleep(time.Second)
			_, _ = writer.Write([]byte(`{"ret":"abc"}`))
		}),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		}}).ListenAndServe())
}

func loadFromHttp(ctx context.Context) ([]byte, error) {
	resp, err := http.Get("http://localhost:9999")
	if err != nil {
		return nil, err
	}
	// nolint
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func NewCacheW() cachew.CacheW {
	redis, err := redisx.NewRedis(&iredis.RedisConfig{})
	if err != nil {
		panic(redis)
	}
	return cachew.NewCacheW(redistorew.NewIRedisStoreW(redis), func(option *cachew.Option) {
		option.AsyncLoad = true
		option.KeyPrefix = "xxxx:"
	})
}

func TestLoad(t *testing.T) {
	const (
		addr = ":9999"
		key  = "test"
	)

	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	go startServer(ctx, addr)

	// 实例化
	cache := NewCacheW()
	// 删除key
	if err := cache.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}

	// http服务为源
	var r ret
	err := cache.Load(ctx, key, 0, &r, func(ctx context.Context) (interface{}, error) {
		return loadFromHttp(ctx)
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Ret != "abc" {
		t.Fatalf("load error")
	}

	// 其他自定义源
	var r2 ret
	err = cache.Load(ctx, key, 0, &r2, func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Second)
		return &ret{Ret: "abc2"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if r2.Ret != "abc2" {
		t.Fatalf("load error")
	}
}
