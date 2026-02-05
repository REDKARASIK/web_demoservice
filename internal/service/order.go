package service

import (
	"context"
	"fmt"
	"log/slog"
	"web_demoservice/internal/cache"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.OrderWithInformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
	GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error)
}

type OrderService struct {
	repo  OrderRepository
	cache *cache.Cache
}

func (s *OrderService) CreateOrder(ctx context.Context, order domain.OrderWithInformation) error {
	err := s.repo.Create(ctx, order)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	if ord, ok := s.cache.Get(id); ok {
		return ord, nil
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.cache.Set(*order)
	return order, nil
}

func (s *OrderService) WarmUp(ctx context.Context) error {
	orders, err := s.repo.GetAllLast24Hours(ctx)
	if err != nil {
		return fmt.Errorf("repo get all: %w", err)
	}

	for _, ord := range orders {
		s.cache.Set(ord)
	}

	slog.Info("Cache warm-up finished", slog.Int("count", len(orders)))
	return nil
}
