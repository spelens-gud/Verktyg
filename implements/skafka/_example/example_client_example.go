package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/IBM/sarama"

	"github.com/spelens-gud/Verktyg/implements/skafka"
	"github.com/spelens-gud/Verktyg/interfaces/ikafka"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

var (
	address = []string{"localhost:9092"}
	ctx     = context.Background()

	testTopic = "skafka-test"
)

// 自定义生产者回调处理
// 同步生产者会阻塞同步执行
// 异步生产者会异步回调执行
type customProduceHandler struct{}

func (c customProduceHandler) OnSuccess(ctx context.Context, message *sarama.ProducerMessage) {
	logger.FromContext(ctx).Infof("produce success: %s key: %s", message.Topic, message.Key)
}

func (c customProduceHandler) OnError(ctx context.Context, err error) {
	logger.FromContext(ctx).Errorf("produce error: %v", err)
}

var _ ikafka.ProducerCallbackHandler = &customProduceHandler{}

// 消费组处理器
type exampleConsumerGroupHandler struct{}

func (e exampleConsumerGroupHandler) OnError(err error) {}

func (e exampleConsumerGroupHandler) OnSessionSetup(session sarama.ConsumerGroupSession) error {
	logger.FromContext(ctx).Infof("session setup : %v", session.MemberID())
	return nil
}

func (e exampleConsumerGroupHandler) OnSessionCleanup(session sarama.ConsumerGroupSession) error {
	logger.FromContext(ctx).Infof("session clean up : %v", session.MemberID())
	return nil
}

func (e exampleConsumerGroupHandler) OnMessage(ctx context.Context, session ikafka.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	logger.FromContext(ctx).Infof("topic: %s ,data: [ %s : %s ]", msg.Topic, msg.Key, msg.Value)
	return nil
}

// sarama消费组处理器
type exampleSaramaConsumerGroupHandler struct{}

func (e exampleSaramaConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	logger.FromContext(ctx).Infof("session setup : %v", session.MemberID())
	return nil
}

func (e exampleSaramaConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	logger.FromContext(ctx).Infof("session clean up : %v", session.MemberID())
	return nil
}

func (e exampleSaramaConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		ctx := session.Context()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-claim.Messages():
			if !ok {
				return errors.New("message channel closed")
			}
			logger.FromContext(ctx).Infof("sarama topic: %s ,data: [ %s : %s ]", msg.Topic, msg.Key, msg.Value)
		}
	}
}

var _ sarama.ConsumerGroupHandler = &exampleSaramaConsumerGroupHandler{}
var _ ikafka.ConsumerGroupHandler = &exampleConsumerGroupHandler{}

type errHandler struct{}

func (e errHandler) OnError(err error) {}

var _ ikafka.ErrorHandler = &errHandler{}

var counter int64

func produce(producer ikafka.Producer) {
	t := time.NewTicker(time.Second).C
	for {
		select {
		case <-ctx.Done():
			return
		case <-t:
			for i := 0; i < 10000; i++ {
				atomic.AddInt64(&counter, 1)
				c := int(atomic.LoadInt64(&counter))
				_ = producer.SendMessage(ctx, &sarama.ProducerMessage{
					Topic: testTopic,
					Key:   sarama.StringEncoder("key" + strconv.Itoa(c)),
					Value: sarama.StringEncoder("value"),
				})
			}
		}
	}
}

func main() {
	//  新建调用端
	client, err := skafka.NewClient(address, func(config *sarama.Config) {
		config.Version = sarama.V2_2_0_0
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	//  新建同步生产者
	syncProducer, err := client.NewSyncProducer(ikafka.WithProducerCallbackHandler(customProduceHandler{}))

	if err != nil {
		panic(err)
	}
	defer syncProducer.Close()

	go produce(syncProducer)

	// 新建异步生产者
	// 使用自定义回调处理器
	asyncProducer, err := client.NewAsyncProducer(ikafka.WithProducerCallbackHandler(customProduceHandler{}))

	if err != nil {
		panic(err)
	}
	defer asyncProducer.Close()

	go produce(asyncProducer)

	//  新建消费者
	consumer, err := client.NewConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(testTopic, 0, 0)
	if err != nil {
		panic(err)
	}
	defer partitionConsumer.Close()

	// 新建消费者组
	consumerGroup, err := client.NewConsumerGroup("groupID")
	if err != nil {
		panic(err)
	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(ctx)
	// cancel退出消费循环
	defer cancel()

	// 使用ikafka定义的handler
	go consumerGroup.StartConsumeWithHandler(ctx, []string{testTopic}, &exampleConsumerGroupHandler{})

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, os.Interrupt)
	<-exit
}
