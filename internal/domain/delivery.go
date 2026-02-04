package domain

import "github.com/google/uuid"

type Delivery struct {
	ID      int64     `db:"id"`
	OrderID uuid.UUID `db:"order_id"`
	Name    string    `db:"name"`
	Phone   string    `db:"phone"`
	Zip     string    `db:"zip"`
	City    string    `db:"city"`
	Address string    `db:"address"`
	Region  *string   `db:"region"`
	Email   string    `db:"email"`
}
