package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	cache2 "web_demoservice/internal/cache"
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
	"github.com/rs/cors"
)

type App struct {
	Router *http.Handler
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

	// Cache
	cache := cache2.NewCache(config.HTTP.CacheTTL)
	cache.StartDeleting(ctx)

	// repo
	orderRepo := repository.NewOrderPostgresRepository(pool)

	// service
	orderService := service.NewOrderService(cache, orderRepo)
	if err = orderService.WarmUp(ctx); err != nil {
		slog.Error("failed to warm up cache", slog.Any("error", err))
	}

	// handler
	orderHandler := handlers.NewOrderHandler(orderService)
	consumerHandler := kafka2.NewOrderHandler(consumer, orderService)
	go consumerHandler.Run(ctx)

	// mux register
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(middleware.PanicCover)
	routs.RegisterOrderRoutes(apiRouter, orderHandler)

	fileServer := http.FileServer(http.Dir("./web"))
	router.PathPrefix("/").Handler(fileServer)

	// Настройка CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept"},
		AllowCredentials: true,
	})

	// Оборачиваем роутер в CORS middleware
	handler := c.Handler(router)

	return &App{
		Router: &handler,
	}, nil
}
