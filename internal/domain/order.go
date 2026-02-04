package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                uuid.UUID `db:"order_id"`
	TrackNumber       string    `db:"track_number"`
	Entry             string    `db:"entry"`
	Locale            string    `db:"locale"`
	InternalSignature *string   `db:"internal_signature"`
	CustomerID        string    `db:"customer_id"`
	DeliveryService   *string   `db:"delivery_service"`
	ShardKey          string    `db:"shard_key"`
	SmID              *int64    `db:"sm_id"`
	DateCreated       time.Time `db:"date_created"`
	OofShard          string    `db:"oof_shard"`
}

type OrderWithItems struct {
	Order
	Items []Item
}

type OrderWithInformation struct {
	OrderWithItems
	Delivery Delivery
	Payment  PaymentWithBank
}
