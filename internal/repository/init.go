package repository

import "github.com/jackc/pgx/v5/pgxpool"

func NewOrderPostgresRepository(pool *pgxpool.Pool) *OrderPostgresRepository {
	return &OrderPostgresRepository{
		pool,
	}
}
