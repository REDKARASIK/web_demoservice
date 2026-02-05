package app

import (
	"context"
	"fmt"
	"web_demoservice/internal/config"
	"web_demoservice/internal/infra/kafka"
	"web_demoservice/internal/infra/postgres"
	"web_demoservice/internal/middleware"
	"web_demoservice/internal/repository"
	"web_demoservice/internal/service"
	"web_demoservice/internal/transport/http/v1/handlers"
	routs "web_demoservice/internal/transport/http/v1/router"
	kafka2 "web_demoservice/internal/transport/kafka"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func NewApp(ctx context.Context, config *config.Config) (*App, error) {
	// postgres
	pool, err := postgres.NewPostgresPool(config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// Kafka
	consumer, err := kafka.NewConsumer(config.Kafka.Brokers, config.Kafka.GroupID, config.Kafka.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	// repo
	orderRepo := repository.NewOrderPostgresRepository(pool)

	// service
	orderService := service.NewOrderService(orderRepo)

	// handler
	orderHandler := handlers.NewOrderHandler(orderService)
	consumerHandler := kafka2.NewOrderHandler(consumer, orderService)
	go consumerHandler.Run(ctx)

	// mux register
	router := mux.NewRouter()
	router = router.PathPrefix("/api/v1").Subrouter()
	router.Use(middleware.PanicCover)
	routs.RegisterOrderRoutes(router, orderHandler)

	return &App{
		Router: router,
	}, nil
}
