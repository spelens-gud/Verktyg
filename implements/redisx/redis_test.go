package redisx

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/spelens-gud/Verktyg.git/interfaces/iredis"
)

func TestRedisCancel(t *testing.T) {
	cli, err := NewRedis(&iredis.RedisConfig{
		DB:         3,
		Addr:       "localhost:6379",
		BaseConfig: iredis.BaseConfig{},
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	// nolint
	defer cli.Close()

	ctx, cf := context.WithCancel(context.Background())
	t.Log(cli.GetRedis(ctx).Set(ctx, "xxx", "dddd", 0).Err())
	t.Log(cli.GetRedis(ctx).Get(ctx, "xxx2").Err())
	cf()
	t.Log(cli.GetRedis(ctx).Get(ctx, "xxx").Err())
}

func TestRedis(t *testing.T) {
	cli, err := NewRedis(&iredis.RedisConfig{
		DB:         3,
		Addr:       "localhost:6379",
		BaseConfig: iredis.BaseConfig{},
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	// nolint
	defer cli.Close()

	eg := errgroup.Group{}
	redis := cli.GetRedis(context.Background())
	for i := 0; i < 100; i++ {
		eg.Go(func() error {
			err = redis.LPush(context.Background(), "dffhd", "sgadasdg").Err()
			if err != nil {
				return err
			}
			_, err = redis.LPop(context.Background(), "dffhd").Result()
			return err
		})
	}
	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkRedis(b *testing.B) {
	var cli, err = NewRedis(&iredis.RedisConfig{
		Addr:       "localhost:6379",
		BaseConfig: iredis.BaseConfig{},
		DB:         3,
	})
	if err != nil {
		b.Fatalf("%+v", err)
	}
	redis := cli.GetRedis(context.Background())
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = redis.LPush(context.Background(), "dffhd", "sgadasdg").Err()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestRedisCluster(t *testing.T) {
	cli, err := NewRedisCluster(&iredis.RedisClusterConfig{
		BaseConfig:   iredis.BaseConfig{MaxRetries: 0},
		Addrs:        []string{"1.2.54.6:8087", "1.2.54.6:8082", "1.2.54.6:8084"},
		MaxRedirects: 8,
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	// nolint
	defer cli.Close()

	eg := errgroup.Group{}
	redis := cli.GetRedis(context.Background())
	for i := 0; i < 1000; i++ {
		eg.Go(func() error {
			err = redis.Set(context.Background(), "gassags", "sgadasdg", 0).Err()
			if err != nil {
				return err
			}
			err = redis.Get(context.Background(), "gassags").Err()
			return err
		})
	}
	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}
