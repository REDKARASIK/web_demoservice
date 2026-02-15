package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"web_demoservice/internal/domain"
	"web_demoservice/internal/infra/kafka"
	"web_demoservice/internal/telemetry"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
			recordCtx, span := otel.Tracer("kafka").Start(ctx, "kafka.consume")
			span.SetAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination", record.Topic),
				attribute.Int("messaging.kafka.partition", int(record.Partition)),
				attribute.Int64("messaging.kafka.offset", record.Offset),
			)
			defer span.End()

			var kafkaDTO OrderKafkaDTO
			if err := json.Unmarshal(record.Value, &kafkaDTO); err != nil {
				wrappedErr := fmt.Errorf("unmarshal kafka record: %w", err)
				slog.Error("failed to unmarshal kafka record", slog.Any("error", wrappedErr))
				span.RecordError(wrappedErr)
				span.SetStatus(codes.Error, wrappedErr.Error())
				telemetry.IncKafkaResult("invalid")
				h.sendToDLQ(recordCtx, record, wrappedErr)
				return
			}

			if err := kafkaDTO.Validate(); err != nil {
				wrappedErr := fmt.Errorf("validate kafka dto: %w", err)
				slog.Error("failed to validate kafka dto", slog.Any("error", wrappedErr))
				span.RecordError(wrappedErr)
				span.SetStatus(codes.Error, wrappedErr.Error())
				telemetry.IncKafkaResult("invalid")
				h.sendToDLQ(recordCtx, record, wrappedErr)
				return
			}

			order, err := kafkaDTO.ToDomain()
			if err != nil {
				wrappedErr := fmt.Errorf("map kafka dto to domain: %w", err)
				slog.Error("failed to map kafka dto to domain", slog.Any("error", wrappedErr))
				span.RecordError(wrappedErr)
				span.SetStatus(codes.Error, wrappedErr.Error())
				telemetry.IncKafkaResult("invalid")
				h.sendToDLQ(recordCtx, record, wrappedErr)
				return
			}

			if err := h.service.CreateOrder(recordCtx, order); err != nil {
				wrappedErr := fmt.Errorf("save order from kafka: %w", err)
				slog.Error("failed to save order from kafka", slog.Any("error", wrappedErr))
				span.RecordError(wrappedErr)
				span.SetStatus(codes.Error, wrappedErr.Error())
				telemetry.IncKafkaResult("error")
				h.sendToDLQ(recordCtx, record, wrappedErr)
				return
			}

			telemetry.IncKafkaResult("ok")
		})
	}
}

func (h *OrderHandler) sendToDLQ(ctx context.Context, record *kgo.Record, cause error) {
	if h.dlq == nil {
		slog.Error("dlq producer is nil", slog.Any("error", cause))
		return
	}

	if err := h.dlq.Publish(ctx, record, cause); err != nil {
		telemetry.IncKafkaDLQPublishFailure()
		slog.Error("failed to publish to dlq", slog.Any("error", err))
	}
}
