package repository

import "github.com/jackc/pgx/v5/pgxpool"

func NewOrderPostgresRepository(db *pgxpool.Pool) *OrderPostgresRepository {
	return &OrderPostgresRepository{
		db: db,
	}
}
