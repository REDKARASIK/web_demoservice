package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"web_demoservice/internal/db"
	"web_demoservice/internal/kafka"
)

func getenv(k, def string) string {
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
	log.SetOutput(file)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	go func() {
		cfg := kafka.Config{
			Brokers:  []string{getenv("KAFKA_BROKERS", "localhost:9092")},
			GroupID:  getenv("KAFKA_GROUP_ID", "web_demoservice-group"),
			Topics:   []string{getenv("KAFKA_TOPIC", "orders-events")},
			ClientID: "web_demoservice-consumer",
		}
		if err := kafka.RunConsumer(ctx, cfg, kafka.HandleLog()); err != nil {
			log.Fatalf("ERROR: kafka: %v", err)
		}
	}()
	pool, err := db.ConnectToPool(ctx)
	if err != nil {
		log.Printf("ERROR: failed to connect to db (%v)\n", err)
	}
	log.Println("INFO: connected to db")
	fmt.Println("INFO: connected to db")
	defer pool.Close()
	<-ctx.Done()
	log.Println("INFO: shutdown")
}
