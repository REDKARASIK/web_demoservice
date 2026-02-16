package cache

import (
	"context"
	"testing"
	"time"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

func TestCache_Get_Miss(t *testing.T) {
	c := NewCache(time.Minute)
	if _, ok := c.Get(context.Background(), uuid.New()); ok {
		t.Fatalf("expected cache miss")
	}
}

func TestCache_SetGet_Hit(t *testing.T) {
	c := NewCache(time.Minute)
	id := uuid.New()
	order := sampleOrder(id)

	c.Set(context.Background(), id, order)
	got, ok := c.Get(context.Background(), id)
	if !ok {
		t.Fatalf("expected cache hit")
	}
	if got == nil || got.ID != id {
		t.Fatalf("expected order %s, got %+v", id, got)
	}
}

func TestCache_Expiration(t *testing.T) {
	ttl := 60 * time.Millisecond
	c := NewCache(ttl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	c.StartDeleting(ctx)

	id := uuid.New()
	order := sampleOrder(id)
	c.Set(context.Background(), id, order)

	time.Sleep(ttl / 2)
	if _, ok := c.Get(context.Background(), id); !ok {
		t.Fatalf("expected cache hit before expiration")
	}

	time.Sleep(ttl + ttl/2)
	if _, ok := c.Get(context.Background(), id); ok {
		t.Fatalf("expected cache entry to expire")
	}
}

func sampleOrder(id uuid.UUID) domain.OrderWithInformation {
	internalSignature := "sig"
	deliveryService := "delivery"
	smID := int64(7)
	region := "region"
	requestID := "req"
	chrtID := int64(1001)
	sale := int64(10)
	size := "M"

	return domain.OrderWithInformation{
		OrderWithItems: domain.OrderWithItems{
			Order: domain.Order{
				ID:                id,
				TrackNumber:       "TRACK",
				Entry:             "WBIL",
				Locale:            "en",
				InternalSignature: &internalSignature,
				CustomerID:        "customer",
				DeliveryService:   &deliveryService,
				ShardKey:          "9",
				SmID:              &smID,
				DateCreated:       time.Now().UTC(),
				OofShard:          "1",
			},
			Items: []domain.Item{
				{
					ChrtID:      &chrtID,
					TrackNumber: "TRACK",
					Price:       100,
					RID:         "RID",
					Name:        "Item",
					Sale:        &sale,
					Size:        &size,
					TotalPrice:  90,
					NmID:        123,
					Brand:       "Brand",
					Status:      1,
				},
			},
		},
		Delivery: domain.Delivery{
			Name:    "Name",
			Phone:   "+1000",
			Zip:     "12345",
			City:    "City",
			Address: "Street",
			Region:  &region,
			Email:   "mail@test.com",
		},
		Payment: domain.PaymentWithBank{
			Payment: domain.Payment{
				Transaction:  "TX",
				RequestID:    &requestID,
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       100,
				PaymentDt:    time.Now().Unix(),
				DeliveryCost: 10,
				GoodsTotal:   90,
				CustomFee:    0,
			},
			Bank: domain.Bank{Name: "bank"},
		},
	}
}
