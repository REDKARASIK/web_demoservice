//go:build integration
// +build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOrderPostgresRepository_CreateAndGet(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set")
	}

	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	var table *string
	if err := pool.QueryRow(ctx, "SELECT to_regclass('orders.orders')").Scan(&table); err != nil {
		t.Fatalf("check schema: %v", err)
	}
	if table == nil {
		t.Skip("schema not migrated")
	}

	repo := NewOrderPostgresRepository(pool)
	order := sampleOrder(uuid.New())

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM orders.order_items WHERE order_id = $1", order.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM orders.payments WHERE order_id = $1", order.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM orders.delivery WHERE order_id = $1", order.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM orders.orders WHERE order_id = $1", order.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM orders.items WHERE rid = $1", order.Items[0].RID)
		_, _ = pool.Exec(ctx, "DELETE FROM banks.banks WHERE name = $1", order.Payment.Bank.Name)
	})

	if err := repo.Create(ctx, order); err != nil {
		t.Fatalf("create order: %v", err)
	}

	got, err := repo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}

	if got.ID != order.ID {
		t.Fatalf("expected id %s, got %s", order.ID, got.ID)
	}
	if got.TrackNumber != order.TrackNumber {
		t.Fatalf("expected track number %s, got %s", order.TrackNumber, got.TrackNumber)
	}
	if got.Payment.Bank.Name != order.Payment.Bank.Name {
		t.Fatalf("expected bank %s, got %s", order.Payment.Bank.Name, got.Payment.Bank.Name)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
	if got.Items[0].RID != order.Items[0].RID {
		t.Fatalf("expected item rid %s, got %s", order.Items[0].RID, got.Items[0].RID)
	}

	orders, err := repo.GetAllLast24Hours(ctx)
	if err != nil {
		t.Fatalf("get all last 24 hours: %v", err)
	}
	found := false
	for _, o := range orders {
		if o.ID == order.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected order in last 24 hours list")
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
				TrackNumber:       "TRACK-" + id.String()[:8],
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
					TrackNumber: "TRACK-" + id.String()[:8],
					Price:       100,
					RID:         "RID-" + id.String()[:8],
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
				Transaction:  "TX-" + id.String()[:8],
				RequestID:    &requestID,
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       100,
				PaymentDt:    time.Now().Unix(),
				DeliveryCost: 10,
				GoodsTotal:   90,
				CustomFee:    0,
			},
			Bank: domain.Bank{Name: "bank-" + id.String()[:8]},
		},
	}
}
