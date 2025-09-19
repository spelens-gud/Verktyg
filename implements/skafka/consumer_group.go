package skafka

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"

	"git.bestfulfill.tech/devops/go-core/interfaces/ikafka"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

func (c *client) NewConsumerGroup(groupID string) (p ikafka.ConsumerGroup, err error) {
	group, err := sarama.NewConsumerGroupFromClient(groupID, c.client)
	if err != nil {
		err = errors.Wrap(err, "new consumer group error")
		return
	}

	return &consumerGroup{
		group:  group,
		config: c.config,
		closed: make(chan struct{}),
	}, nil
}

type consumerGroup struct {
	config *ikafka.ClientConfig
	group  sarama.ConsumerGroup
	closed chan struct{}
}

func defaultOnError(err error) {
	logger.FromBackground().WithTag(ikafka.ErrLogTag).Errorf("consumer error: %v", err)
}

func (c *consumerGroup) Close() (err error) {
	close(c.closed)
	return c.group.Close()
}

func (c *consumerGroup) StartConsumeWithHandler(ctx context.Context, topics []string, handler ikafka.ConsumerGroupHandler) {
	var onErr func(err error)
	if errHandler, ok := handler.(ikafka.ErrorHandler); ok {
		onErr = errHandler.OnError
	} else {
		onErr = defaultOnError
	}
	c.StartConsumeWithSaramaHandler(ctx, topics, &consumerGroupHandler{handler: handler, config: c.config}, onErr)
}

func (c *consumerGroup) StartConsumeWithSaramaHandler(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler, onError func(error)) {
	if c.config.Consumer.Return.Errors {
		go func() {
			select {
			case <-c.closed:
				return
			case err, ok := <-c.group.Errors():
				if ok {
					onError(err)
				}
			}
		}()
	}

	for {
		select {
		case <-c.closed:
			logger.FromContext(ctx).Infof("consume stop by group closed")
			return
		case <-ctx.Done():
			logger.FromContext(ctx).Infof("consume stop by parent context")
			return
		default:
			if err := c.group.Consume(ctx, topics, handler); err != nil {
				logger.FromContext(ctx).Errorf("group consume error: %v", err)
			}
			time.Sleep(time.Second)
		}
	}
}

type consumerGroupHandler struct {
	handler ikafka.ConsumerGroupHandler
	config  *ikafka.ClientConfig
}

func (d consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return d.handler.OnSessionSetup(session)
}

func (d consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return d.handler.OnSessionCleanup(session)
}

func (d consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		ctx := session.Context()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-claim.Messages():
			if msg != nil {
				doConsume(ctx, d.config.Address[0], "group", msg, func(ctx context.Context) error {
					return d.handler.OnMessage(ctx, ikafka.NewConsumerGroupSession(session, claim), msg)
				})
			}
		}
	}
}

type consumerGroupFuncHandler struct {
	onSessionSetup   func(session sarama.ConsumerGroupSession) error
	onSessionCleanup func(session sarama.ConsumerGroupSession) error
	onMessage        func(ctx context.Context, session ikafka.ConsumerGroupSession, msg *sarama.ConsumerMessage) error
}

func (c consumerGroupFuncHandler) OnSessionSetup(session sarama.ConsumerGroupSession) error {
	if c.onSessionSetup == nil {
		return nil
	}
	return c.onSessionSetup(session)
}

func (c consumerGroupFuncHandler) OnSessionCleanup(session sarama.ConsumerGroupSession) error {
	if c.onSessionCleanup == nil {
		return nil
	}
	return c.onSessionCleanup(session)
}

func (c consumerGroupFuncHandler) OnMessage(ctx context.Context, session ikafka.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	if c.onMessage == nil {
		return nil
	}
	return c.onMessage(ctx, session, msg)
}
