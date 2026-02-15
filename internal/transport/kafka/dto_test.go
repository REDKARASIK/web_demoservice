package kafka

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOrderKafkaDTO_ToDomain_OptionalFieldsNil(t *testing.T) {
	dto := validDTO()

	got, err := dto.ToDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.InternalSignature != nil {
		t.Fatalf("expected InternalSignature nil")
	}
	if got.DeliveryService != nil {
		t.Fatalf("expected DeliveryService nil")
	}
	if got.SmID != nil {
		t.Fatalf("expected SmID nil")
	}
	if got.Delivery.Region != nil {
		t.Fatalf("expected Delivery.Region nil")
	}
	if got.Payment.RequestID != nil {
		t.Fatalf("expected Payment.RequestID nil")
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
	if got.Items[0].ChrtID != nil || got.Items[0].Sale != nil || got.Items[0].Size != nil {
		t.Fatalf("expected item optional fields nil")
	}
}

func TestOrderKafkaDTO_ToDomain_InvalidUUID(t *testing.T) {
	dto := OrderKafkaDTO{
		OrderUID: "bad-uuid",
	}

	if _, err := dto.ToDomain(); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestOrderKafkaDTO_Validate_OK(t *testing.T) {
	dto := validDTO()
	if err := dto.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrderKafkaDTO_Validate_MissingRequired(t *testing.T) {
	dto := validDTO()
	dto.OrderUID = ""
	dto.TrackNumber = ""
	dto.Delivery.Email = ""
	dto.Payment.Transaction = ""
	dto.Items = nil

	if err := dto.Validate(); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func validDTO() OrderKafkaDTO {
	return OrderKafkaDTO{
		OrderUID:    uuid.New().String(),
		TrackNumber: "TRACK",
		Entry:       "WBIL",
		Locale:      "en",
		CustomerID:  "customer",
		ShardKey:    "9",
		DateCreated: time.Now().UTC(),
		OofShard:    "1",
		Delivery: DeliveryDTO{
			Name:    "Name",
			Phone:   "+1000",
			Zip:     "12345",
			City:    "City",
			Address: "Street",
			Email:   "mail@test.com",
		},
		Payment: PaymentDTO{
			Transaction:  "TX",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       100,
			PaymentDt:    time.Now().Unix(),
			Bank:         "bank",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Items: []ItemDTO{
			{
				TrackNumber: "TRACK",
				Price:       100,
				RID:         "RID",
				Name:        "Item",
				TotalPrice:  90,
				NmID:        123,
				Brand:       "Brand",
				Status:      1,
			},
		},
	}
}
