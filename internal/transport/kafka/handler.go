package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"web_demoservice/internal/domain"
	"web_demoservice/internal/infra/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.OrderWithInformation) error
}

type OrderHandler struct {
	consumer *kafka.Consumer
	service  OrderService
}

func NewOrderHandler(consumer *kafka.Consumer, service OrderService) *OrderHandler {
	return &OrderHandler{
		consumer: consumer,
		service:  service,
	}
}

func (h *OrderHandler) Run(ctx context.Context) {
	for {
		fetches := h.consumer.Fetch(ctx)
		fetches.EachRecord(func(record *kgo.Record) {
			var kafkaDTO OrderKafkaDTO
			if err := json.Unmarshal(record.Value, &kafkaDTO); err != nil {
				slog.Error("failed to unmarshal kafka record", slog.Any("error", err))
				return
			}

			order, err := kafkaDTO.ToDomain()
			if err != nil {
				slog.Error("failed to map kafka dto to domain", slog.Any("error", err))
				return
			}

			if err := h.service.CreateOrder(ctx, order); err != nil {
				slog.Error("failed to save order from kafka", slog.Any("error", err))
			}
		})
	}
}
