package ikafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type (
	ProducerOpt struct {
		CallbackHandler ProducerCallbackHandler
	}

	ProducerOption func(option *ProducerOpt)

	ProducerOptions []ProducerOption

	ProducerCallbackHandler interface {
		OnSuccess(ctx context.Context, message *sarama.ProducerMessage)
		OnError(ctx context.Context, err error)
	}
)

const ErrLogTag = "KAFKA_ERR"

func WithProducerCallbackHandler(h ProducerCallbackHandler) ProducerOption {
	return func(option *ProducerOpt) {
		if h == nil {
			panic("invalid callback handler")
		}
		option.CallbackHandler = h
	}
}

type DefaultCallbackHandler struct{}

var defaultCallbackHandler = &DefaultCallbackHandler{}

func (d DefaultCallbackHandler) OnSuccess(ctx context.Context, message *sarama.ProducerMessage) {}

func (d DefaultCallbackHandler) OnError(ctx context.Context, err error) {
	lg := logger.FromContext(ctx).WithTag(ErrLogTag)
	if pErr, ok := err.(*sarama.ProducerError); ok {
		msg := pErr.Msg
		lg.WithFields(map[string]interface{}{
			"value": fmt.Sprintf("%s", msg.Value),
			"key":   fmt.Sprintf("%s", msg.Key),
			"topic": msg.Topic,
		})
	}
	lg.Errorf("produce message error: %v", err)
}

func (opts ProducerOptions) Init() *ProducerOpt {
	o := &ProducerOpt{
		CallbackHandler: defaultCallbackHandler,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
