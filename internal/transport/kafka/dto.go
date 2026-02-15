package kafka

import (
	"errors"
	"fmt"
	"strings"
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
	InternalSignature *string     `json:"internal_signature,omitempty"`
	CustomerID        string      `json:"customer_id"`
	DeliveryService   *string     `json:"delivery_service,omitempty"`
	ShardKey          string      `json:"shardkey"`
	SmID              *int64      `json:"sm_id,omitempty"`
	DateCreated       time.Time   `json:"date_created"`
	OofShard          string      `json:"oof_shard"`
}

type DeliveryDTO struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Zip     string  `json:"zip"`
	City    string  `json:"city"`
	Address string  `json:"address"`
	Region  *string `json:"region,omitempty"`
	Email   string  `json:"email"`
}

type PaymentDTO struct {
	Transaction  string  `json:"transaction"`
	RequestID    *string `json:"request_id,omitempty"`
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
	ChrtID      *int64  `json:"chrt_id,omitempty"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	RID         string  `json:"rid"`
	Name        string  `json:"name"`
	Sale        *int64  `json:"sale,omitempty"`
	Size        *string `json:"size,omitempty"`
	TotalPrice  float64 `json:"total_price"`
	NmID        int64   `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}

func (d *OrderKafkaDTO) ToDomain() (domain.OrderWithInformation, error) {
	uid, err := uuid.Parse(d.OrderUID)
	if err != nil {
		return domain.OrderWithInformation{}, fmt.Errorf("parse order_uid: %w", err)
	}

	items := make([]domain.Item, len(d.Items))
	for i, it := range d.Items {
		items[i] = domain.Item{
			ChrtID:      it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			RID:         it.RID,
			Name:        it.Name,
			Sale:        it.Sale,
			Size:        it.Size,
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
				InternalSignature: d.InternalSignature,
				CustomerID:        d.CustomerID,
				DeliveryService:   d.DeliveryService,
				ShardKey:          d.ShardKey,
				SmID:              d.SmID,
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
			Region:  d.Delivery.Region,
			Email:   d.Delivery.Email,
		},
		Payment: domain.PaymentWithBank{
			Payment: domain.Payment{
				Transaction:  d.Payment.Transaction,
				RequestID:    d.Payment.RequestID,
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

func (d *OrderKafkaDTO) Validate() error {
	var errs []error

	if isBlank(d.OrderUID) {
		errs = append(errs, errors.New("order_uid is required"))
	} else if _, err := uuid.Parse(d.OrderUID); err != nil {
		errs = append(errs, fmt.Errorf("order_uid invalid: %w", err))
	}

	if isBlank(d.TrackNumber) {
		errs = append(errs, errors.New("track_number is required"))
	}
	if isBlank(d.Entry) {
		errs = append(errs, errors.New("entry is required"))
	}
	if isBlank(d.Locale) {
		errs = append(errs, errors.New("locale is required"))
	}
	if isBlank(d.CustomerID) {
		errs = append(errs, errors.New("customer_id is required"))
	}
	if isBlank(d.ShardKey) {
		errs = append(errs, errors.New("shardkey is required"))
	}
	if isBlank(d.OofShard) {
		errs = append(errs, errors.New("oof_shard is required"))
	}
	if d.DateCreated.IsZero() {
		errs = append(errs, errors.New("date_created is required"))
	}

	if isBlank(d.Delivery.Name) {
		errs = append(errs, errors.New("delivery.name is required"))
	}
	if isBlank(d.Delivery.Phone) {
		errs = append(errs, errors.New("delivery.phone is required"))
	}
	if isBlank(d.Delivery.Zip) {
		errs = append(errs, errors.New("delivery.zip is required"))
	}
	if isBlank(d.Delivery.City) {
		errs = append(errs, errors.New("delivery.city is required"))
	}
	if isBlank(d.Delivery.Address) {
		errs = append(errs, errors.New("delivery.address is required"))
	}
	if isBlank(d.Delivery.Email) {
		errs = append(errs, errors.New("delivery.email is required"))
	}

	if isBlank(d.Payment.Transaction) {
		errs = append(errs, errors.New("payment.transaction is required"))
	}
	if isBlank(d.Payment.Currency) {
		errs = append(errs, errors.New("payment.currency is required"))
	}
	if isBlank(d.Payment.Provider) {
		errs = append(errs, errors.New("payment.provider is required"))
	}
	if isBlank(d.Payment.Bank) {
		errs = append(errs, errors.New("payment.bank is required"))
	}
	if d.Payment.PaymentDt <= 0 {
		errs = append(errs, errors.New("payment.payment_dt must be positive"))
	}
	if d.Payment.Amount < 0 || d.Payment.DeliveryCost < 0 || d.Payment.GoodsTotal < 0 || d.Payment.CustomFee < 0 {
		errs = append(errs, errors.New("payment values must be non-negative"))
	}

	if len(d.Items) == 0 {
		errs = append(errs, errors.New("items must not be empty"))
	}
	for i, it := range d.Items {
		prefix := fmt.Sprintf("items[%d].", i)
		if isBlank(it.TrackNumber) {
			errs = append(errs, errors.New(prefix+"track_number is required"))
		}
		if isBlank(it.RID) {
			errs = append(errs, errors.New(prefix+"rid is required"))
		}
		if isBlank(it.Name) {
			errs = append(errs, errors.New(prefix+"name is required"))
		}
		if isBlank(it.Brand) {
			errs = append(errs, errors.New(prefix+"brand is required"))
		}
		if it.NmID <= 0 {
			errs = append(errs, errors.New(prefix+"nm_id must be positive"))
		}
		if it.Price < 0 || it.TotalPrice < 0 {
			errs = append(errs, errors.New(prefix+"price/total_price must be non-negative"))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}
