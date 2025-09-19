package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"git.bestfulfill.tech/devops/go-core/implements/skafka"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
	"git.bestfulfill.tech/devops/go-core/kits/kruntime"
)

type Order struct {
	ID        int
	Money     int
	Timestamp time.Time
}

const (
	Topic        = "test_topic"
	ConsumeGroup = "test_group"
	Endpoint     = "10.43.3.11:9092"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	client, err := skafka.NewClient([]string{Endpoint})
	if err != nil {
		return
	}
	// nolint
	defer client.Close()

	// 初始化消费方法
	consumeFunc, err := client.Consume(ConsumeGroup, []string{Topic}, func(ctx context.Context, message []byte) (err error) {
		var o Order
		if err = json.Unmarshal(message, &o); err != nil {
			return
		}
		logger.FromContext(ctx).Infof("%v", o)
		return
	})
	if err != nil {
		return
	}

	// 开启协程消费
	go consumeFunc(ctx)

	// 生产
	go func() {
		for range time.NewTicker(time.Second).C {
			_ = client.SyncSendMessageX(ctx, Topic, "key", Order{
				ID:        rand.Intn(1000),
				Money:     rand.Intn(1000),
				Timestamp: time.Now(),
			})
		}
	}()

	<-kruntime.StopChan()
}
