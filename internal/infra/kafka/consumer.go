package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	client *kgo.Client
}

func NewConsumer(brokers []string, groupID, topic string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic))

	return &Consumer{client: client}, err
}

func (c *Consumer) Fetch(ctx context.Context) kgo.Fetches {
	return c.client.PollFetches(ctx)
}
