package skafka

import (
	"context"
	"net/textproto"

	"github.com/IBM/sarama"
	shopifySarama "github.com/Shopify/sarama"

	"github.com/spelens-gud/Verktyg/interfaces/itrace"
	"github.com/spelens-gud/Verktyg/kits/ktrace/tracer"
)

func startProducerTrace(ctx context.Context, address, topic, typ string) (sp itrace.Span, header []sarama.RecordHeader) {
	tmp := make(map[string][]string)
	sp = tracer.InjectMetadata(ctx, "KAFKA_PRODUCE:"+topic, address, tmp)
	itrace.SetMqTag(sp, address, topic, "kafka", "sarama", typ, true)
	return sp, parseHeader(tmp)
}

func startConsumerTrace(ctx context.Context, message *sarama.ConsumerMessage, address, topic, typ string) (itrace.Span, context.Context) {
	m := parseFromHeader(message.Headers)
	sp, ctx := tracer.ExtractMetadata(ctx, "KAFKA_CONSUME:"+topic, m)
	itrace.SetMqTag(sp, address, topic, "kafka", "sarama", typ, false)
	return sp, ctx
}

// startConsumerTraceCluster 处理 cluster 消费的追踪（使用 Shopify/sarama 类型）
func startConsumerTraceCluster(ctx context.Context, message *shopifySarama.ConsumerMessage, address, topic, typ string) (itrace.Span, context.Context) {
	m := parseFromHeaderCluster(message.Headers)
	sp, ctx := tracer.ExtractMetadata(ctx, "KAFKA_CONSUME:"+topic, m)
	itrace.SetMqTag(sp, address, topic, "kafka", "sarama", typ, false)
	return sp, ctx
}

func parseFromHeader(header []*sarama.RecordHeader) (m textproto.MIMEHeader) {
	if len(header) == 0 {
		return nil
	}
	m = make(textproto.MIMEHeader, len(header))
	for _, h := range header {
		m.Set(string(h.Key), string(h.Value))
	}
	return m
}

// parseFromHeaderCluster 解析 cluster 消息头（使用 Shopify/sarama 类型）
func parseFromHeaderCluster(header []*shopifySarama.RecordHeader) (m textproto.MIMEHeader) {
	if len(header) == 0 {
		return nil
	}
	m = make(textproto.MIMEHeader, len(header))
	for _, h := range header {
		m.Set(string(h.Key), string(h.Value))
	}
	return m
}

func parseHeader(m textproto.MIMEHeader) (header []sarama.RecordHeader) {
	if len(m) == 0 {
		return nil
	}
	header = make([]sarama.RecordHeader, 0, len(m))
	for k := range m {
		header = append(header, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(m.Get(k)),
		})
	}
	return header
}
