package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"

	"git.bestfulfill.tech/devops/go-core/implements/skafka"
	"git.bestfulfill.tech/devops/go-core/implements/worker"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
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

	// 初始化客户端
	client, err := skafka.NewClient([]string{Endpoint})
	if err != nil {
		return
	}
	defer client.Close()

	// 初始化管理框架
	manager := worker.NewWorkerManager(func(option *worker.Option) {
		option.CronOptions = append(option.CronOptions, cron.WithSeconds())
	})

	// 注册定时生产
	manager.RegisterCron("*/5 * * * * *", func(ctx context.Context) (err error) {
		return client.AsyncSendMessageX(ctx, Topic, "key", Order{
			ID:        rand.Intn(100),
			Money:     rand.Intn(100),
			Timestamp: time.Now(),
		})
	})

	// 注册消费任务 写法1
	consumeFunc, err := client.Consume(ConsumeGroup, []string{Topic}, func(ctx context.Context, message []byte) error {
		logger.FromContext(ctx).WithTag("CONSUME_1").Infof("%s", message)
		return nil
	})
	if err != nil {
		return
	}
	manager.RegisterTasks(consumeFunc)

	// 注册消费任务 写法2  Must 自动Panic错误处理
	manager.MustRegisterTask(client.Consume(ConsumeGroup+"_2", []string{Topic}, func(ctx context.Context, message []byte) error {
		logger.FromContext(ctx).WithTag("CONSUME_2").Infof("%s", message)
		return nil
	}))

	// 启动进程管理
	manager.Start()

}
