package kafka

import (
	"time"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

type OrderKafkaDTO struct {
	OrderUID          string      `json:"order_uid"`
	TrackNumber       string      `json:"track_number"`
	Entry             string      `json:"entry"`
	Delivery          DeliveryDTO `json:"delivery"`
	Payment           PaymentDTO  `json:"payment"`
	Items             []ItemDTO   `json:"items"`
	Locale            string      `json:"locale"`
	InternalSignature string      `json:"internal_signature"`
	CustomerID        string      `json:"customer_id"`
	DeliveryService   string      `json:"delivery_service"`
	ShardKey          string      `json:"shardkey"`
	SmID              int64       `json:"sm_id"`
	DateCreated       time.Time   `json:"date_created"`
	OofShard          string      `json:"oof_shard"`
}

type DeliveryDTO struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type PaymentDTO struct {
	Transaction  string  `json:"transaction"`
	RequestID    string  `json:"request_id"`
	Currency     string  `json:"currency"`
	Provider     string  `json:"provider"`
	Amount       float64 `json:"amount"`
	PaymentDt    int64   `json:"payment_dt"`
	Bank         string  `json:"bank"`
	DeliveryCost float64 `json:"delivery_cost"`
	GoodsTotal   int64   `json:"goods_total"`
	CustomFee    float64 `json:"custom_fee"`
}

type ItemDTO struct {
	ChrtID      int64   `json:"chrt_id"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	RID         string  `json:"rid"`
	Name        string  `json:"name"`
	Sale        int64   `json:"sale"`
	Size        string  `json:"size"`
	TotalPrice  float64 `json:"total_price"`
	NmID        int64   `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}

func (d *OrderKafkaDTO) ToDomain() (domain.OrderWithInformation, error) {
	uid, err := uuid.Parse(d.OrderUID)
	if err != nil {
		return domain.OrderWithInformation{}, err
	}

	items := make([]domain.Item, len(d.Items))
	for i, it := range d.Items {
		items[i] = domain.Item{
			ChrtID:      &it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			RID:         it.RID,
			Name:        it.Name,
			Sale:        &it.Sale,
			Size:        &it.Size,
			TotalPrice:  it.TotalPrice,
			NmID:        it.NmID,
			Brand:       it.Brand,
			Status:      it.Status,
		}
	}

	return domain.OrderWithInformation{
		OrderWithItems: domain.OrderWithItems{
			Order: domain.Order{
				ID:                uid,
				TrackNumber:       d.TrackNumber,
				Entry:             d.Entry,
				Locale:            d.Locale,
				InternalSignature: &d.InternalSignature,
				CustomerID:        d.CustomerID,
				DeliveryService:   &d.DeliveryService,
				ShardKey:          d.ShardKey,
				SmID:              &d.SmID,
				DateCreated:       d.DateCreated,
				OofShard:          d.OofShard,
			},
			Items: items,
		},
		Delivery: domain.Delivery{
			Name:    d.Delivery.Name,
			Phone:   d.Delivery.Phone,
			Zip:     d.Delivery.Zip,
			City:    d.Delivery.City,
			Address: d.Delivery.Address,
			Region:  &d.Delivery.Region,
			Email:   d.Delivery.Email,
		},
		Payment: domain.PaymentWithBank{
			Payment: domain.Payment{
				Transaction:  d.Payment.Transaction,
				RequestID:    &d.Payment.RequestID,
				Currency:     d.Payment.Currency,
				Provider:     d.Payment.Provider,
				Amount:       d.Payment.Amount,
				PaymentDt:    d.Payment.PaymentDt,
				DeliveryCost: d.Payment.DeliveryCost,
				GoodsTotal:   d.Payment.GoodsTotal,
				CustomFee:    d.Payment.CustomFee,
			},
			Bank: domain.Bank{Name: d.Payment.Bank},
		},
	}, nil
}
