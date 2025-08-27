package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	api2 "web_demoservice/internal/api"
	"web_demoservice/internal/cache"
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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("INFO: APP started")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.ConnectToPool(ctx)
	if err != nil {
		log.Fatalf("ERROR: failed to connect to db: %v", err)
	}
	defer pool.Close()
	log.Println("INFO: connected to db")

	orderRepo := repo.NewOrderRepo(pool)
	ordersCache := cache.NewOrdersCache()
	api := api2.NewServer(ordersCache, orderRepo)
	fs := http.FileServer(http.Dir("web"))
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.Handle("/api/", http.StripPrefix("/api", api))
	httpAddr := getEnv("HTTP_ADDR", ":8081")
	go func() {
		log.Printf("INFO: http listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ERROR: http server: %v", err)
		}
	}()
	cfg := kafka.Config{
		Brokers:  []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		GroupID:  getEnv("KAFKA_GROUP_ID", "web_demoservice-group"),
		Topics:   []string{getEnv("KAFKA_TOPIC", "orders-events")},
		ClientID: "web_demoservice-consumer",
	}
	go func() {
		handler := kafka.NewOrderHandler(orderRepo)
		if err := kafka.RunConsumer(ctx, cfg, handler); err != nil {
			log.Fatalf("ERROR: kafka consumer stopped: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("INFO: shutdown")
}
