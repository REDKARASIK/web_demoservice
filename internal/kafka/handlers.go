package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"web_demoservice/internal/model"
	"web_demoservice/internal/repo"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewOrderHandler(repo repo.OrderRepository) MessageHandler {
	return func(ctx context.Context, rec *kgo.Record) error {
		var order model.Order
		if err := json.Unmarshal(rec.Value, &order); err != nil {
			log.Printf("[WARN]: bad json: %v", err)
			return nil
		}
		if order.OrderUID == "" {
			return errors.New("empty order_uid")
		}
		return repo.UpsertOrder(ctx, order)
	}
}
