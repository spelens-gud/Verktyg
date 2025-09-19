package skafka

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"

	"git.bestfulfill.tech/devops/go-core/interfaces/ikafka"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

var _ ikafka.ClusterConsumeHandler = &exampleClusterConsumeHandler{}

type exampleClusterConsumeHandler struct{}

func (e exampleClusterConsumeHandler) OnError(err error) {
	logger.FromBackground().Errorf("error : %v", err)
	// 处理错误
}

func (e exampleClusterConsumeHandler) OnMessage(ctx context.Context, consumer *cluster.Consumer, message *sarama.ConsumerMessage) (err error) {
	// 处理消息
	logger.FromContext(ctx).Infof("get message: %s key: %s", message.Topic, message.Key)
	return nil
}

func (e exampleClusterConsumeHandler) OnNotifications(notification *cluster.Notification) {
	// 处理事件
}

func TestCluster(t *testing.T) {
	clusterClient, err := NewClusterClient([]string{"localhost:9092"}, func(config *cluster.Config) {
		// 自定义配置
		config.Version = sarama.V2_2_0_0
		config.Group.Return.Notifications = true
	})
	if err != nil {
		t.Fatal(err)
	}
	// nolint
	defer clusterClient.Close()
	consumer, err := clusterClient.NewConsumer("groupID", []string{"skafka-test"})
	if err != nil {
		t.Fatal(err)
	}
	// nolint
	defer consumer.Close()
	// 使用处理器消费
	ctx, cancel := context.WithCancel(context.Background())
	consumer.ConsumeWithHandler(ctx, &exampleClusterConsumeHandler{})
	defer cancel()
	select {}
}
