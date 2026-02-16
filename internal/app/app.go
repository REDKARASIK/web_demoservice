package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	cache2 "web_demoservice/internal/cache"
	"web_demoservice/internal/config"
	"web_demoservice/internal/infra/kafka"
	"web_demoservice/internal/infra/postgres"
	"web_demoservice/internal/middleware"
	"web_demoservice/internal/repository"
	"web_demoservice/internal/service"
	"web_demoservice/internal/telemetry"
	"web_demoservice/internal/transport/http/v1/handlers"
	routs "web_demoservice/internal/transport/http/v1/router"
	kafka2 "web_demoservice/internal/transport/kafka"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type App struct {
	Router *http.Handler
}

func NewApp(ctx context.Context, config *config.Config) (*App, error) {
	// postgres
	pool, err := postgres.NewPostgresPool(config.DB, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// Kafka
	consumer, err := kafka.NewConsumer(config.Kafka.Brokers, config.Kafka.GroupID, config.Kafka.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}
	dlqProducer, err := kafka.NewProducer(config.Kafka.Brokers, config.Kafka.DLQTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka dlq producer: %w", err)
	}

	// Cache
	cache := cache2.NewCache(config.HTTP.CacheTTL)
	cache.StartDeleting(ctx)
	cacheObs := telemetry.WrapCache(cache)

	// repo
	orderRepo := repository.NewOrderPostgresRepository(pool)
	repoObs := telemetry.WrapOrderRepository(orderRepo)
	if config.Metrics.Enabled {
		startRepositoryPing(ctx, repoObs, config.DB.HealthCheckPeriod)
	}

	// service
	orderService := service.NewOrderService(repoObs, cacheObs)
	orderServiceObs := telemetry.WrapOrderService(orderService)
	if err = orderServiceObs.WarmUp(ctx); err != nil {
		slog.Error("failed to warm up cache", slog.Any("error", err))
	}

	// handler
	orderHandler := handlers.NewOrderHandler(orderServiceObs)
	orderHandlerObs := handlers.NewLoggingOrderHandler(orderHandler)
	consumerHandler := kafka2.NewOrderHandler(consumer, dlqProducer, orderServiceObs)
	go consumerHandler.Run(ctx)

	// mux register
	router := mux.NewRouter()
	if config.Telemetry.Enabled {
		router.Use(otelhttp.NewMiddleware("http_server"))
	}
	if config.Metrics.Enabled {
		router.Use(telemetry.MetricsMiddleware)
		metricsPath := config.Metrics.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		router.Handle(metricsPath, telemetry.MetricsHandler())
	}
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(middleware.PanicCover)
	routs.RegisterOrderRoutes(apiRouter, orderHandlerObs)

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

type repositoryPinger interface {
	Ping(ctx context.Context) error
}

func startRepositoryPing(ctx context.Context, repo repositoryPinger, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		if err := repo.Ping(ctx); err != nil {
			slog.Warn("repository ping failed", slog.Any("error", err))
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := repo.Ping(ctx); err != nil {
					slog.Warn("repository ping failed", slog.Any("error", err))
				}
			}
		}
	}()
}
