package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client recordProducer
	topic  string
}

type recordProducer interface {
	Produce(ctx context.Context, record *kgo.Record, fn func(*kgo.Record, error))
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	return newProducerWithClient(client, topic)
}

func newProducerWithClient(client recordProducer, topic string) (*Producer, error) {
	if topic == "" {
		return nil, fmt.Errorf("dlq topic is required")
	}
	if client == nil {
		return nil, fmt.Errorf("producer client is required")
	}

	return &Producer{client: client, topic: topic}, nil
}

func (p *Producer) Publish(ctx context.Context, src *kgo.Record, cause error) error {
	if src == nil {
		return fmt.Errorf("nil source record")
	}
	if cause == nil {
		return fmt.Errorf("nil error")
	}

	headers := make([]kgo.RecordHeader, 0, len(src.Headers)+5)
	headers = append(headers, src.Headers...)
	headers = append(headers,
		kgo.RecordHeader{Key: "dlq_error", Value: []byte(cause.Error())},
		kgo.RecordHeader{Key: "dlq_source_topic", Value: []byte(src.Topic)},
		kgo.RecordHeader{Key: "dlq_source_partition", Value: []byte(fmt.Sprintf("%d", src.Partition))},
		kgo.RecordHeader{Key: "dlq_source_offset", Value: []byte(fmt.Sprintf("%d", src.Offset))},
		kgo.RecordHeader{Key: "dlq_ts", Value: []byte(time.Now().UTC().Format(time.RFC3339Nano))},
	)

	record := &kgo.Record{
		Topic:   p.topic,
		Key:     src.Key,
		Value:   src.Value,
		Headers: headers,
	}

	errCh := make(chan error, 1)
	p.client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		errCh <- err
	})
	if err := <-errCh; err != nil {
		return fmt.Errorf("produce dlq: %w", err)
	}

	return nil
}
