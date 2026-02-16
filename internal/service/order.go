package service

import (
	"context"
	"fmt"
	"log/slog"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.OrderWithInformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
	GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error)
}

type Cache interface {
	Set(ctx context.Context, key uuid.UUID, value domain.OrderWithInformation)
	Get(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool)
}

func NewOrderService(repo OrderRepository, cache Cache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

type OrderService struct {
	cache Cache
	repo  OrderRepository
}

func (s *OrderService) CreateOrder(ctx context.Context, order domain.OrderWithInformation) error {
	err := s.repo.Create(ctx, order)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	if ord, ok := s.cache.Get(ctx, id); ok {
		return ord, nil
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	s.cache.Set(ctx, (*order).ID, *order)
	return order, nil
}

func (s *OrderService) WarmUp(ctx context.Context) error {
	orders, err := s.repo.GetAllLast24Hours(ctx)
	if err != nil {
		return fmt.Errorf("repo get all: %w", err)
	}

	for _, ord := range orders {
		s.cache.Set(ctx, ord.ID, ord)
	}

	slog.Info("Cache warm-up finished", slog.Int("count", len(orders)))
	return nil
}
