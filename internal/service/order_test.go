package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

type mockOrderRepo struct {
	createFn          func(ctx context.Context, order domain.OrderWithInformation) error
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
	getAllLast24Hours func(ctx context.Context) ([]domain.OrderWithInformation, error)
	createCalls       int
	getByIDCalls      int
	getAllLast24Calls int
}

func (m *mockOrderRepo) Create(ctx context.Context, order domain.OrderWithInformation) error {
	m.createCalls++
	if m.createFn != nil {
		return m.createFn(ctx, order)
	}
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	m.getByIDCalls++
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockOrderRepo) GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error) {
	m.getAllLast24Calls++
	if m.getAllLast24Hours != nil {
		return m.getAllLast24Hours(ctx)
	}
	return nil, nil
}

type mockCache struct {
	getFn    func(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool)
	setFn    func(ctx context.Context, key uuid.UUID, value domain.OrderWithInformation)
	getCalls int
	setCalls int
}

func (m *mockCache) Set(ctx context.Context, key uuid.UUID, value domain.OrderWithInformation) {
	m.setCalls++
	if m.setFn != nil {
		m.setFn(ctx, key, value)
	}
}

func (m *mockCache) Get(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool) {
	m.getCalls++
	if m.getFn != nil {
		return m.getFn(ctx, key)
	}
	return nil, false
}

var _ Cache = (*mockCache)(nil)

func TestOrderService_CreateOrder_PropagatesError(t *testing.T) {
	wantErr := errors.New("repo error")
	repo := &mockOrderRepo{
		createFn: func(ctx context.Context, order domain.OrderWithInformation) error {
			return wantErr
		},
	}
	cache := &mockCache{}

	svc := NewOrderService(repo, cache)
	if err := svc.CreateOrder(context.Background(), sampleOrder(uuid.New())); !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestOrderService_GetOrder_FromCache(t *testing.T) {
	id := uuid.New()
	order := sampleOrder(id)

	repo := &mockOrderRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			t.Fatalf("repo should not be called on cache hit")
			return nil, nil
		},
	}
	cache := &mockCache{
		getFn: func(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool) {
			return &order, true
		},
	}

	svc := NewOrderService(repo, cache)
	got, err := svc.GetOrder(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != order.ID {
		t.Fatalf("expected cached order, got %+v", got)
	}
	if cache.getCalls != 1 {
		t.Fatalf("expected cache Get called once, got %d", cache.getCalls)
	}
	if repo.getByIDCalls != 0 {
		t.Fatalf("expected repo GetByID not called, got %d", repo.getByIDCalls)
	}
}

func TestOrderService_GetOrder_FromRepo_SetsCache(t *testing.T) {
	id := uuid.New()
	order := sampleOrder(id)

	repo := &mockOrderRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			return &order, nil
		},
	}
	cache := &mockCache{
		getFn: func(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool) {
			return nil, false
		},
		setFn: func(ctx context.Context, key uuid.UUID, value domain.OrderWithInformation) {
			if key != id {
				t.Fatalf("expected cache Set key %s, got %s", id, key)
			}
		},
	}

	svc := NewOrderService(repo, cache)
	got, err := svc.GetOrder(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != order.ID {
		t.Fatalf("expected repo order, got %+v", got)
	}
	if cache.setCalls != 1 {
		t.Fatalf("expected cache Set called once, got %d", cache.setCalls)
	}
	if repo.getByIDCalls != 1 {
		t.Fatalf("expected repo GetByID called once, got %d", repo.getByIDCalls)
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
