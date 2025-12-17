package skafka

import (
	"sync"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"

	"github.com/spelens-gud/Verktyg.git/interfaces/ikafka"
)

type client struct {
	client sarama.Client
	config *ikafka.ClientConfig

	asyncProducer     ikafka.Producer
	asyncProducerInit sync.Once

	syncProducer     ikafka.Producer
	syncProducerInit sync.Once
}

func NewClient(address []string, configs ...ikafka.ClientOption) (cli ikafka.Client, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "new client error")
		}
	}()

	if len(address) == 0 {
		err = errors.New("invalid address")
		return
	}

	config := &ikafka.ClientConfig{
		Address: address,
		Config:  sarama.NewConfig(),
	}

	config.Init(configs...)

	c, err := sarama.NewClient(config.Address, config.Config)
	if err != nil {
		return
	}

	return &client{
		client: c,
		config: config,
	}, nil
}

func (c *client) Close() error {
	if c.syncProducer != nil {
		_ = c.syncProducer.Close()
	}
	if c.asyncProducer != nil {
		_ = c.asyncProducer.Close()
	}
	return c.client.Close()
}

func (c *client) Client() sarama.Client {
	return c.client
}

func (c *client) NewConsumer() (consumer ikafka.Consumer, err error) {
	consumer, err = sarama.NewConsumerFromClient(c.client)
	if err != nil {
		err = errors.Wrap(err, "new consumer error")
		return
	}
	return consumer, nil
}

func (c *client) NewSyncProducer(opts ...ikafka.ProducerOption) (p ikafka.Producer, err error) {
	producer, err := sarama.NewSyncProducerFromClient(c.client)
	if err != nil {
		err = errors.Wrap(err, "new async producer error")
		return
	}

	opt := ikafka.ProducerOptions(opts).Init()

	return &syncProducer{config: c.config, SyncProducer: producer, opt: opt, client: c.client}, nil
}

func (c *client) NewAsyncProducer(opts ...ikafka.ProducerOption) (p ikafka.Producer, err error) {
	producer, err := sarama.NewAsyncProducerFromClient(c.client)
	if err != nil {
		err = errors.Wrap(err, "new async producer error")
		return
	}

	var (
		opt       = ikafka.ProducerOptions(opts).Init()
		aProducer = &asyncProducer{config: c.config, client: c.client, AsyncProducer: producer, opt: opt, closed: make(chan struct{})}
	)

	aProducer.startAsyncDaemon()

	return aProducer, nil
}
