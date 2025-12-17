package skafka

import (
	cluster "github.com/bsm/sarama-cluster"
	"github.com/pkg/errors"

	"github.com/spelens-gud/Verktyg/interfaces/ikafka"
)

type clusterClient struct {
	client *cluster.Client
	Config ikafka.ClusterConfig
}

func (c *clusterClient) Close() error {
	return c.client.Close()
}

func NewClusterClient(address []string, opts ...ikafka.ClusterOption) (client ikafka.ClusterClient, err error) {
	if len(address) == 0 {
		err = errors.New("invalid address")
		return
	}

	config := ikafka.ClusterConfig{
		Config:  cluster.NewConfig(),
		Address: address,
	}

	config.Init(opts...)

	clusterCli, err := cluster.NewClient(address, config.Config)
	if err != nil {
		err = errors.Wrap(err, "new cluster client error")
		return
	}
	return &clusterClient{
		client: clusterCli,
		Config: config,
	}, nil
}

func (c *clusterClient) Client() *cluster.Client {
	return c.client
}

func (c *clusterClient) NewConsumer(groupID string, topics []string) (consumer ikafka.ClusterConsumer, err error) {
	cs, err := cluster.NewConsumerFromClient(c.client, groupID, topics)
	if err != nil {
		err = errors.Wrap(err, "new cluster consumer error")
		return
	}
	return &clusterConsumer{
		config:   c.Config,
		consumer: cs,
		closed:   make(chan struct{}),
	}, nil
}
