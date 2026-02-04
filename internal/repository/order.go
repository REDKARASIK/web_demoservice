package repository

import (
	"context"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderPostgresRepository struct {
	db *pgxpool.Pool
}

func (r *OrderPostgresRepository) Create(ctx context.Context, order domain.OrderWithInformation) error {
	panic("implement me")
}

func (r *OrderPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	panic("implement me")
}
