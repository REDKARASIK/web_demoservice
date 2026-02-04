package domain

import "github.com/google/uuid"

type Payment struct {
	ID           int64     `db:"id"`
	OrderID      uuid.UUID `db:"order_id"`
	Transaction  string    `db:"transaction"`
	RequestID    *string   `db:"request_id"`
	Currency     string    `db:"currency"`
	Provider     string    `db:"provider"`
	Amount       float64   `db:"amount"`
	PaymentDt    int64     `db:"payment_dt"`
	BankID       int64     `db:"bank_id"`
	DeliveryCost float64   `db:"delivery_cost"`
	GoodsTotal   int64     `db:"goods_total"`
	CustomFee    float64   `db:"custom_fee"`
}

type PaymentWithBank struct {
	Payment
	Bank Bank
}
