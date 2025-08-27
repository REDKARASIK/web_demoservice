package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"web_demoservice/internal/db"
	"web_demoservice/internal/kafka"
	"web_demoservice/internal/repo"
)

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("INFO: APP started")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.ConnectToPool(ctx)
	if err != nil {
		log.Fatalf("ERROR: failed to connect to db (%v)\n", err)
	}
	defer pool.Close()
	fmt.Println("INFO: connected to db")

	cfg := kafka.Config{
		Brokers:  []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		GroupID:  getEnv("KAFKA_GROUP_ID", "web_demoservice-group"),
		Topics:   []string{getEnv("KAFKA_TOPIC", "orders-events")},
		ClientID: "web_demoservice-consumer",
	}
	orderRepo := repo.NewOrderRepo(pool)
	handler := kafka.NewOrderHandler(orderRepo)
	if err := kafka.RunConsumer(ctx, cfg, handler); err != nil {
		log.Fatalf("ERROR: kafka consumer stopped with error: %v", err)
	}

	log.Println("INFO: shutdown")
}
