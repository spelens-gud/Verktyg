package ikafka

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/kits/knet"
)

type Client interface {
	Close() error

	Client() sarama.Client

	SyncSendMessageX(ctx context.Context, topic, key string, message interface{}) (err error)
	AsyncSendMessageX(ctx context.Context, topic, key string, message interface{}) (err error)

	SyncSendMessage(ctx context.Context, message *sarama.ProducerMessage) (err error)
	AsyncSendMessage(ctx context.Context, message *sarama.ProducerMessage) (err error)

	// 快速创建消费任务
	Consume(groupID string, topics []string, handleFunc func(ctx context.Context, message []byte) error) (f func(ctx context.Context), err error)
	// 新建定制消费者
	NewConsumer() (consumer Consumer, err error)
	NewConsumerGroup(groupID string) (p ConsumerGroup, err error)
	// 新建生产者
	NewSyncProducer(opts ...ProducerOption) (p Producer, err error)
	NewAsyncProducer(opts ...ProducerOption) (p Producer, err error)
}

// 旧版本cluster消费方式 建议采用新版consumer group
type ClusterClient interface {
	Close() error

	Client() *cluster.Client
	// 新建定制消费者
	Consume(groupID string, topics []string, handleFunc func(ctx context.Context, message []byte) error) (f func(ctx context.Context), err error)
	// 新建定制消费者
	NewConsumer(groupID string, topics []string) (consumer ClusterConsumer, err error)
}

type ClusterConsumer interface {
	Close() error

	Consumer() *cluster.Consumer
	ConsumeWithHandler(ctx context.Context, handler ClusterConsumeHandler)
}

type (
	ClientConfig struct {
		*sarama.Config
		Address []string
	}

	ClusterConfig struct {
		*cluster.Config
		Address []string
	}

	Producer interface {
		Close() error

		// 快速封装方法
		SendMessageX(ctx context.Context, topic, key string, value interface{}) (err error)
		SendMessagesX(ctx context.Context, topic, key string, value ...interface{}) (err error)
		// async producer返回err如无意外都为nil
		SendMessage(ctx context.Context, message *sarama.ProducerMessage) (err error)
		SendMessages(ctx context.Context, message ...*sarama.ProducerMessage) (err error)
	}

	ConsumerGroup interface {
		Close() error

		StartConsumeWithHandler(ctx context.Context, topics []string, handler ConsumerGroupHandler)
	}

	Consumer interface {
		sarama.Consumer
	}

	ClientOption  func(config *sarama.Config)
	ClusterOption func(config *cluster.Config)

	ClusterConsumeHandler interface {
		ErrorHandler
		OnMessage(ctx context.Context, consumer *cluster.Consumer, message *sarama.ConsumerMessage) error
		OnNotifications(notification *cluster.Notification)
	}
)

type consumerGroupSession struct {
	sarama.ConsumerGroupSession
	claim sarama.ConsumerGroupClaim
}

type ConsumerGroupHandler interface {
	OnSessionSetup(session sarama.ConsumerGroupSession) error
	OnSessionCleanup(session sarama.ConsumerGroupSession) error
	OnMessage(ctx context.Context, session ConsumerGroupSession, msg *sarama.ConsumerMessage) error
}

type ErrorHandler interface {
	OnError(err error)
}

type ConsumerGroupSession interface {
	sarama.ConsumerGroupSession
	Claim() sarama.ConsumerGroupClaim
}

func (c consumerGroupSession) Claim() sarama.ConsumerGroupClaim {
	return c.claim
}

func NewConsumerGroupSession(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) ConsumerGroupSession {
	return &consumerGroupSession{
		ConsumerGroupSession: session,
		claim:                claim,
	}
}

func (cfg *ClientConfig) Init(configs ...ClientOption) {
	initSaramaConfig(cfg.Config)
	for _, c := range configs {
		c(cfg.Config)
	}
}

func (cfg *ClusterConfig) Init(configs ...ClusterOption) {
	initSaramaConfig(&cfg.Config.Config)
	for _, c := range configs {
		c(cfg.Config)
	}
}

func initSaramaConfig(cfg *sarama.Config) {
	// 连接池参数
	if cfg.Net.DialTimeout == 0 {
		cfg.Net.DialTimeout = time.Second * 5
	}
	if cfg.Net.KeepAlive == 0 {
		cfg.Net.KeepAlive = time.Second * 120
	}
	cfg.Metadata.RefreshFrequency = time.Minute * 3
	// 客户端ID
	cfg.ClientID = iconfig.GetApplicationNameSpace() + "_" + iconfig.GetApplicationAppName()
	// tcp连接封装
	cfg.Net.Proxy.Enable = true
	cfg.Net.Proxy.Dialer = knet.WrapTcpDialer(cfg.Net.DialTimeout, cfg.Net.KeepAlive, nil)
	// 默认版本号
	cfg.Version = sarama.V0_11_0_2
	// 开启异步生产回调
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = true
	// 修复版本兼容Interval
	cfg.Consumer.Offsets.CommitInterval = cfg.Consumer.Offsets.AutoCommit.Interval
}
