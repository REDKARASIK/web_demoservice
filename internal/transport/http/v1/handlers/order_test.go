package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type mockOrderService struct {
	getOrderFn func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
}

func (m *mockOrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	return m.getOrderFn(ctx, id)
}

func TestOrderHandler_GetOrder_BadRequestOnMissingID(t *testing.T) {
	h := NewOrderHandler(&mockOrderService{
		getOrderFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			t.Fatalf("service should not be called")
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/order/", nil)
	req = mux.SetURLVars(req, map[string]string{})
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOrderHandler_GetOrder_BadRequestOnInvalidUUID(t *testing.T) {
	h := NewOrderHandler(&mockOrderService{
		getOrderFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			t.Fatalf("service should not be called")
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/order/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"order_id": "bad"})
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	h := NewOrderHandler(&mockOrderService{
		getOrderFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			return nil, pgx.ErrNoRows
		},
	})

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/order/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"order_id": id})
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestOrderHandler_GetOrder_InternalError(t *testing.T) {
	h := NewOrderHandler(&mockOrderService{
		getOrderFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			return nil, errors.New("boom")
		},
	})

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/order/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"order_id": id})
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestOrderHandler_GetOrder_OK(t *testing.T) {
	order := sampleOrder(uuid.New())
	h := NewOrderHandler(&mockOrderService{
		getOrderFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
			return &order, nil
		},
	})

	id := order.ID.String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/order/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"order_id": id})
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["order_uid"] != order.ID.String() {
		t.Fatalf("expected order_uid %s, got %v", order.ID.String(), got["order_uid"])
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
