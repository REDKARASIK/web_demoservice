package service

import (
	"context"
	"fmt"
	"log/slog"
	"web_demoservice/internal/domain"
	"web_demoservice/internal/interfaces"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.OrderWithInformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
	GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error)
}

func NewOrderService(repo OrderRepository, cache interfaces.Cache[uuid.UUID, domain.OrderWithInformation]) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

type OrderService struct {
	cache interfaces.Cache[uuid.UUID, domain.OrderWithInformation]
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
	if ord, ok := s.cache.Get(id); ok {
		return ord, nil
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	s.cache.Set((*order).ID, *order)
	return order, nil
}

func (s *OrderService) WarmUp(ctx context.Context) error {
	orders, err := s.repo.GetAllLast24Hours(ctx)
	if err != nil {
		return fmt.Errorf("repo get all: %w", err)
	}

	for _, ord := range orders {
		s.cache.Set(ord.ID, ord)
	}

	slog.Info("Cache warm-up finished", slog.Int("count", len(orders)))
	return nil
}
