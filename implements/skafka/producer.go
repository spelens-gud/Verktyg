package skafka

import (
	"context"
	"io"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/spelens-gud/Verktyg/internal/incontext"

	"github.com/Shopify/sarama"

	"github.com/spelens-gud/Verktyg/interfaces/ikafka"
	"github.com/spelens-gud/Verktyg/interfaces/itrace"
	"github.com/spelens-gud/Verktyg/kits/kdb"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
	"github.com/spelens-gud/Verktyg/kits/ktrace/tracer"
)

type (
	syncProducer struct {
		config *ikafka.ClientConfig
		opt    *ikafka.ProducerOpt
		client sarama.Client
		sarama.SyncProducer
	}

	asyncProducer struct {
		config *ikafka.ClientConfig
		opt    *ikafka.ProducerOpt
		closed chan struct{}
		client sarama.Client
		sarama.AsyncProducer
	}
)

const (
	headerWarnMessage = "kafka header support at least v0.11.0.0,header dropped"

	keyContextTime           = incontext.Key("skafka.time")
	keyContextOriginMetadata = incontext.Key("skafka.meta")
)

func parseMessageBody(value interface{}) (encoder sarama.Encoder, err error) {
	switch v := value.(type) {
	case string:
		encoder = sarama.StringEncoder(v)
	case []byte:
		encoder = sarama.ByteEncoder(v)
	case io.Reader:
		var b []byte
		if b, err = io.ReadAll(v); err != nil {
			return
		}
		encoder = sarama.ByteEncoder(b)
	default:
		var b []byte
		if b, err = jsoniter.Marshal(value); err != nil {
			return
		}
		encoder = sarama.ByteEncoder(b)
	}
	return
}

func (s *syncProducer) Close() error {
	return s.SyncProducer.Close()
}

func (s *syncProducer) SendMessagesX(ctx context.Context, topic, key string, values ...interface{}) (err error) {
	msgs := make([]*sarama.ProducerMessage, 0, len(values))
	for _, v := range values {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(key),
		}
		if msg.Value, err = parseMessageBody(v); err != nil {
			return
		}
		msgs = append(msgs, msg)
	}
	return s.SendMessages(ctx, msgs...)
}

func (s *syncProducer) SendMessageX(ctx context.Context, topic, key string, value interface{}) (err error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
	}
	if msg.Value, err = parseMessageBody(value); err != nil {
		return
	}
	return s.SendMessage(ctx, msg)
}

func (s *syncProducer) SendMessages(ctx context.Context, messages ...*sarama.ProducerMessage) (err error) {
	var errors sarama.ProducerErrors

	total := len(messages)
	retChan := make(chan *sarama.ProducerError, total)
	done := 0

	for _, msg := range messages {
		go func(msg *sarama.ProducerMessage) {
			if e := s.SendMessage(ctx, msg); e != nil {
				retChan <- &sarama.ProducerError{
					Msg: msg,
					Err: e,
				}
			} else {
				retChan <- nil
			}
		}(msg)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case producerError := <-retChan:
			if producerError != nil {
				errors = append(errors, producerError)
			}
			if done += 1; done == total {
				close(retChan)
				if len(errors) == 0 {
					return nil
				}
				return errors
			}
		}
	}
}

func (s *syncProducer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) (err error) {
	var (
		sp, headers = startProducerTrace(ctx, s.config.Address[0], msg.Topic, "sync")
		t           = time.Now()
	)
	defer sp.Finish()

	if headers != nil {
		if s.config.Version.IsAtLeast(sarama.V0_11_0_0) {
			msg.Headers = headers
		} else {
			logger.FromContext(ctx).Warn(headerWarnMessage)
		}
	}

	// 发送消息
	partition, offset, err := s.SyncProducer.SendMessage(msg)

	kdb.KafkaProduceMetrics(s.config.Address[0], msg.Topic, err, time.Since(t))
	itrace.SetMqMessageTag(sp, int(partition), int(offset))

	if err != nil {
		sp.Error(err, logger.FromContext(ctx).FieldData())
		if s.opt.CallbackHandler != nil {
			go s.opt.CallbackHandler.OnError(ctx, err)
		}
	} else {
		if s.opt.CallbackHandler != nil {
			go s.opt.CallbackHandler.OnSuccess(ctx, msg)
		}
	}
	return err
}

func (s *asyncProducer) Close() error {
	close(s.closed)
	return s.AsyncProducer.Close()
}

func (s *asyncProducer) SendMessagesX(ctx context.Context, topic, key string, values ...interface{}) (err error) {
	for _, v := range values {
		_ = s.SendMessageX(ctx, topic, key, v)
	}
	return
}

func (s *asyncProducer) SendMessages(ctx context.Context, msgs ...*sarama.ProducerMessage) (err error) {
	for _, msg := range msgs {
		_ = s.SendMessage(ctx, msg)
	}
	return nil
}

func (s *asyncProducer) SendMessageX(ctx context.Context, topic, key string, value interface{}) (err error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
	}
	if msg.Value, err = parseMessageBody(value); err != nil {
		return
	}
	return s.SendMessage(ctx, msg)
}

func (s *asyncProducer) initMessageTelemetry(ctx context.Context, msg *sarama.ProducerMessage) {
	var (
		endpoint    = s.config.Address[0]
		sp, headers = startProducerTrace(ctx, endpoint, msg.Topic, "async")
		start       = time.Now()
	)

	if s.config.Producer.Return.Successes && s.config.Producer.Return.Errors {
		// 可以通过回调进行遥测
		ctx = tracer.SpanWithContext(ctx, sp)
		ctx = keyContextTime.WithValue(ctx, start)
		if msg.Metadata != nil {
			ctx = keyContextOriginMetadata.WithValue(ctx, msg.Metadata)
		}
		msg.Metadata = ctx
	} else {
		defer func() {
			sp.Finish()
			kdb.KafkaProduceMetrics(endpoint, msg.Topic, nil, time.Since(start))
		}()
	}

	if headers == nil {
		return
	}

	if s.config.Version.IsAtLeast(sarama.V0_11_0_0) {
		msg.Headers = headers
	} else {
		logger.FromContext(ctx).Warn(headerWarnMessage)
	}
}

func (s *asyncProducer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) (err error) {
	s.initMessageTelemetry(ctx, msg)
	// nolint
	s.AsyncProducer.Input() <- msg
	return nil
}

func (s *asyncProducer) startAsyncDaemon() {
	// 成功回调
	if s.config.Producer.Return.Successes {
		go func() {
			for {
				select {
				case <-s.closed:
					return
				// 成功回调
				case msg := <-s.Successes():
					if msg != nil {
						s.produceCallback(msg)
					}
				}
			}
		}()
	}

	// 失败回调
	if s.config.Producer.Return.Errors {
		go func() {
			for {
				select {
				case <-s.closed:
					return
				case e := <-s.Errors():
					if e != nil {
						s.errorCallback(e)
					}
				}
			}
		}()
	}
}

func (s *asyncProducer) produceCallback(msg *sarama.ProducerMessage) {
	// 提取context
	ctx, ok := msg.Metadata.(context.Context)
	if ok {
		if metadata := keyContextOriginMetadata.Value(ctx); metadata != nil {
			msg.Metadata = metadata
		}
		metricsProduce(ctx, s.config.Address[0], msg, nil)
	} else {
		ctx = context.Background()
	}

	if s.opt.CallbackHandler != nil {
		go s.opt.CallbackHandler.OnSuccess(ctx, msg)
	}
}

func (s *asyncProducer) errorCallback(e *sarama.ProducerError) {
	// 提取context
	ctx, ok := e.Msg.Metadata.(context.Context)
	if ok {
		if metadata := keyContextOriginMetadata.Value(ctx); metadata != nil {
			e.Msg.Metadata = metadata
		}
		metricsProduce(ctx, s.config.Address[0], e.Msg, e)
	} else {
		ctx = context.Background()
	}

	if e.Err == sarama.ErrUnknownTopicOrPartition {
		go func() {
			_ = s.client.RefreshMetadata(e.Msg.Topic)
		}()
	}

	if s.opt.CallbackHandler != nil {
		go s.opt.CallbackHandler.OnError(ctx, e)
	}
}

func metricsProduce(ctx context.Context, addr string, msg *sarama.ProducerMessage, err error) {
	// metrics
	if t, ok := keyContextTime.Value(ctx).(time.Time); ok {
		kdb.KafkaProduceMetrics(addr, msg.Topic, err, time.Since(t))
	}

	// 链路追踪
	if sp := tracer.SpanFromContext(ctx); sp != nil {
		if err != nil {
			sp.Error(err, logger.FromContext(ctx).FieldData())
		}
		itrace.SetMqMessageTag(sp, int(msg.Partition), int(msg.Offset))
		sp.Finish()
	}
}
