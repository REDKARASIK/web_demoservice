package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"web_demoservice/internal/domain"
	"web_demoservice/internal/infra/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.OrderWithInformation) error
}

type DLQProducer interface {
	Publish(ctx context.Context, record *kgo.Record, cause error) error
}

type OrderHandler struct {
	consumer *kafka.Consumer
	dlq      DLQProducer
	service  OrderService
}

func NewOrderHandler(consumer *kafka.Consumer, dlq DLQProducer, service OrderService) *OrderHandler {
	return &OrderHandler{
		consumer: consumer,
		dlq:      dlq,
		service:  service,
	}
}

func (h *OrderHandler) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		fetches := h.consumer.Fetch(ctx)
		fetches.EachRecord(func(record *kgo.Record) {
			var kafkaDTO OrderKafkaDTO
			if err := json.Unmarshal(record.Value, &kafkaDTO); err != nil {
				wrappedErr := fmt.Errorf("unmarshal kafka record: %w", err)
				slog.Error("failed to unmarshal kafka record", slog.Any("error", wrappedErr))
				h.sendToDLQ(ctx, record, wrappedErr)
				return
			}

			if err := kafkaDTO.Validate(); err != nil {
				wrappedErr := fmt.Errorf("validate kafka dto: %w", err)
				slog.Error("failed to validate kafka dto", slog.Any("error", wrappedErr))
				h.sendToDLQ(ctx, record, wrappedErr)
				return
			}

			order, err := kafkaDTO.ToDomain()
			if err != nil {
				wrappedErr := fmt.Errorf("map kafka dto to domain: %w", err)
				slog.Error("failed to map kafka dto to domain", slog.Any("error", wrappedErr))
				h.sendToDLQ(ctx, record, wrappedErr)
				return
			}

			if err := h.service.CreateOrder(ctx, order); err != nil {
				wrappedErr := fmt.Errorf("save order from kafka: %w", err)
				slog.Error("failed to save order from kafka", slog.Any("error", wrappedErr))
				h.sendToDLQ(ctx, record, wrappedErr)
			}
		})
	}
}

func (h *OrderHandler) sendToDLQ(ctx context.Context, record *kgo.Record, cause error) {
	if h.dlq == nil {
		slog.Error("dlq producer is nil", slog.Any("error", cause))
		return
	}

	if err := h.dlq.Publish(ctx, record, cause); err != nil {
		slog.Error("failed to publish to dlq", slog.Any("error", err))
	}
}
