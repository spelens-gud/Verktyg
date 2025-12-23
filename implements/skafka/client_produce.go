package skafka

import (
	"context"

	"github.com/IBM/sarama"
)

func (c *client) SyncSendMessageX(ctx context.Context, topic, key string, message interface{}) (err error) {
	c.syncProducerInit.Do(func() {
		c.syncProducer, err = c.NewSyncProducer()
	})
	if err != nil {
		return
	}
	return c.syncProducer.SendMessageX(ctx, topic, key, message)
}

func (c *client) SyncSendMessage(ctx context.Context, message *sarama.ProducerMessage) (err error) {
	c.syncProducerInit.Do(func() {
		c.syncProducer, err = c.NewSyncProducer()
	})
	if err != nil {
		return
	}
	return c.syncProducer.SendMessage(ctx, message)
}

func (c *client) AsyncSendMessageX(ctx context.Context, topic, key string, message interface{}) (err error) {
	c.asyncProducerInit.Do(func() {
		c.asyncProducer, err = c.NewAsyncProducer()
	})
	if err != nil {
		return
	}
	return c.asyncProducer.SendMessageX(ctx, topic, key, message)
}

func (c *client) AsyncSendMessage(ctx context.Context, message *sarama.ProducerMessage) (err error) {
	c.asyncProducerInit.Do(func() {
		c.asyncProducer, err = c.NewAsyncProducer()
	})
	if err != nil {
		return
	}
	return c.asyncProducer.SendMessage(ctx, message)
}
