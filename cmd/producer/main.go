package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"math/rand"
	"strings"
	"time"
	"web_demoservice/internal/transport/kafka"

	"github.com/brianvoe/gofakeit"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	brokers := flag.String("brokers", "localhost:19092", "comma-separated brokers")
	topic := flag.String("topic", "orders", "topic name")
	count := flag.Int("count", 10, "messages to send")
	invalidRate := flag.Float64("invalid-rate", 0.2, "rate of invalid orders")
	flag.Parse()

	gofakeit.Seed(time.Now().UnixNano())
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	client, err := kgo.NewClient(kgo.SeedBrokers(strings.Split(*brokers, ",")...))
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < *count; i++ {
		var payload []byte
		if rng.Float64() < *invalidRate {
			if rng.Intn(2) == 0 {
				payload = []byte("not-json")
			} else {
				dto := makeInvalidOrder()
				payload, _ = json.Marshal(dto)
			}
		} else {
			dto := makeValidOrder()
			payload, _ = json.Marshal(dto)
		}

		record := kgo.Record{Topic: *topic, Value: payload}
		res := client.ProduceSync(ctx, &record)
		if res.FirstErr() != nil {
			slog.Error("producer error", "error", res.FirstErr())
		}
	}
}

func makeValidOrder() kafka.OrderKafkaDTO {
	internalSig := gofakeit.Word()
	deliveryService := gofakeit.Word()
	smID := int64(gofakeit.Number(1, 1000))
	region := gofakeit.State()
	requestID := gofakeit.Word()
	chrtID := int64(gofakeit.Number(1000, 9999))
	sale := int64(gofakeit.Number(0, 50))
	size := gofakeit.RandString([]string{"S", "M", "L"})

	orderID := gofakeit.UUID()
	track := "TRACK-" + gofakeit.UUID()[:8]

	return kafka.OrderKafkaDTO{
		OrderUID:          orderID,
		TrackNumber:       track,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: &internalSig,
		CustomerID:        gofakeit.Username(),
		DeliveryService:   &deliveryService,
		ShardKey:          "9",
		SmID:              &smID,
		DateCreated:       time.Now().UTC(),
		OofShard:          "1",
		Delivery: kafka.DeliveryDTO{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  &region,
			Email:   gofakeit.Email(),
		},
		Payment: kafka.PaymentDTO{
			Transaction:  "TX-" + gofakeit.UUID()[:8],
			RequestID:    &requestID,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       float64(gofakeit.Number(100, 5000)),
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: float64(gofakeit.Number(100, 1000)),
			GoodsTotal:   int64(gofakeit.Number(100, 5000)),
			CustomFee:    0,
		},
		Items: []kafka.ItemDTO{
			{
				ChrtID:      &chrtID,
				TrackNumber: track,
				Price:       float64(gofakeit.Number(100, 2000)),
				RID:         "RID-" + gofakeit.UUID()[:8],
				Name:        gofakeit.Word(),
				Sale:        &sale,
				Size:        &size,
				TotalPrice:  float64(gofakeit.Number(100, 2000)),
				NmID:        int64(gofakeit.Number(100000, 999999)),
				Brand:       "Brand",
				Status:      1,
			},
		},
	}
}

func makeInvalidOrder() kafka.OrderKafkaDTO {
	// Нарушаем обязательные поля: пустые строки, отрицательные суммы, пустые items
	dto := makeValidOrder()
	dto.OrderUID = "bad-uuid"
	dto.TrackNumber = ""
	dto.CustomerID = ""
	dto.Items = nil
	dto.Payment.Amount = -10
	dto.Delivery.Email = ""
	return dto
}
