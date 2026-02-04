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
			var order domain.OrderWithInformation
			if err := json.Unmarshal(record.Value, &order); err != nil {
				slog.Error("failed to unmarshal order", slog.Any("record", record))
				return
			}

			err := h.service.CreateOrder(ctx, order)
			if err != nil {
				slog.Error("failed to create order", slog.Any("order", order))
			}
		})
	}
}
