package skafka

import (
	"context"
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/ikafka"

	shopifySarama "github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"

	"github.com/IBM/sarama"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
	"github.com/spelens-gud/Verktyg/kits/kdb"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

func doConsume(ctx context.Context, address, typ string, msg *sarama.ConsumerMessage, do func(ctx context.Context) error) {
	sp, ctx := startConsumerTrace(ctx, msg, address, msg.Topic, typ)
	itrace.SetMqMessageTag(sp, int(msg.Partition), int(msg.Offset))
	defer func(startTime time.Time, err error) {
		kdb.KafkaConsumeMetrics(address, msg.Topic, err, time.Since(startTime))
		if err != nil {
			sp.Error(err, logger.FromContext(ctx).FieldData())
		}
		sp.Finish()
	}(time.Now(), do(ctx))
}

// doConsumeCluster 处理 cluster 消费（使用 Shopify/sarama 类型）
func doConsumeCluster(ctx context.Context, address, typ string, msg *shopifySarama.ConsumerMessage, do func(ctx context.Context) error) {
	sp, ctx := startConsumerTraceCluster(ctx, msg, address, msg.Topic, typ)
	itrace.SetMqMessageTag(sp, int(msg.Partition), int(msg.Offset))
	defer func(startTime time.Time, err error) {
		kdb.KafkaConsumeMetrics(address, msg.Topic, err, time.Since(startTime))
		if err != nil {
			sp.Error(err, logger.FromContext(ctx).FieldData())
		}
		sp.Finish()
	}(time.Now(), do(ctx))
}

func (c *client) Consume(groupID string, topics []string, handleFunc func(ctx context.Context, message []byte) error) (f func(ctx context.Context), err error) {
	consumer, err := c.NewConsumerGroup(groupID)
	if err != nil {
		return
	}
	return func(ctx context.Context) {
		// nolint
		defer consumer.Close()
		consumer.StartConsumeWithHandler(ctx, topics, &consumerGroupFuncHandler{
			onMessage: func(ctx context.Context, session ikafka.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
				defer func() {
					// mark message
					session.MarkMessage(msg, "")
				}()
				return handleFunc(ctx, msg.Value)
			},
		})
	}, nil
}

func (c *clusterClient) Consume(groupID string, topics []string, handleFunc func(ctx context.Context, message []byte) error) (f func(ctx context.Context), err error) {
	consumer, err := c.NewConsumer(groupID, topics)
	if err != nil {
		return
	}
	return func(ctx context.Context) {
		// nolint
		defer consumer.Close()
		consumer.ConsumeWithHandler(ctx, clusterConsumeFuncHandler{
			onError: defaultOnError,
			onMessages: func(ctx context.Context, consumer *cluster.Consumer, message *shopifySarama.ConsumerMessage) error {
				return handleFunc(ctx, message.Value)
			},
		})
	}, nil
}
