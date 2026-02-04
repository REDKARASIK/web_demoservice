package service

import (
	"context"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

// TODO: add caching

type OrderRepository interface {
	Create(ctx context.Context, order domain.OrderWithInformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
}

type OrderService struct {
	repo OrderRepository
}

func (s *OrderService) CreateOrder(ctx context.Context, order domain.OrderWithInformation) error {
	err := s.repo.Create(ctx, order)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return order, nil
}
