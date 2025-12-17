package skafka

import (
	"context"

	"github.com/Shopify/sarama"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"

	cluster "github.com/bsm/sarama-cluster"

	"github.com/spelens-gud/Verktyg/interfaces/ikafka"
	"github.com/spelens-gud/Verktyg/kits/kdb"
)

type clusterConsumer struct {
	config   ikafka.ClusterConfig
	consumer *cluster.Consumer
	closed   chan struct{}
}

func (consumer *clusterConsumer) Close() error {
	close(consumer.closed)
	return consumer.consumer.Close()
}

func (consumer *clusterConsumer) Consumer() *cluster.Consumer {
	return consumer.consumer
}

func (consumer *clusterConsumer) consumeError(handler ikafka.ClusterConsumeHandler, err error) {
	kdb.KafkaConsumeMetrics(consumer.config.Address[0], "", err, 0)
	go handler.OnError(err)
}

func (consumer *clusterConsumer) ConsumeWithHandler(ctx context.Context, handler ikafka.ClusterConsumeHandler) {
	for {
		select {
		case <-ctx.Done():
			logger.FromContext(ctx).Infof("consume stop by group closed")
			return
		case <-consumer.closed:
			logger.FromContext(ctx).Infof("consume stop by parent context")
			return
		case msg := <-consumer.consumer.Messages():
			if msg != nil {
				doConsume(ctx, consumer.config.Address[0], "cluster", msg, func(ctx context.Context) error {
					return handler.OnMessage(ctx, consumer.consumer, msg)
				})
			}
		case err := <-consumer.consumer.Errors():
			if err != nil {
				consumer.consumeError(handler, err)
			}
		case notify := <-consumer.consumer.Notifications():
			if notify != nil {
				go handler.OnNotifications(notify)
			}
		}
	}
}

var _ ikafka.ClusterConsumeHandler = &clusterConsumeFuncHandler{}

type clusterConsumeFuncHandler struct {
	onError         func(error)
	onMessages      func(context.Context, *cluster.Consumer, *sarama.ConsumerMessage) error
	onNotifications func(*cluster.Notification)
}

func (c clusterConsumeFuncHandler) OnError(err error) {
	if c.onError == nil {
		return
	}
	c.onError(err)
}

func (c clusterConsumeFuncHandler) OnMessage(ctx context.Context, consumer *cluster.Consumer, message *sarama.ConsumerMessage) error {
	if c.onMessages == nil {
		return nil
	}
	return c.onMessages(ctx, consumer, message)
}

func (c clusterConsumeFuncHandler) OnNotifications(notification *cluster.Notification) {
	if c.onNotifications == nil {
		return
	}
	c.onNotifications(notification)
}
