package app

import (
	"context"
	"fmt"
	"net/http"
	"web_demoservice/internal/config"
	"web_demoservice/internal/infra/postgres"
)

type App struct {
	Router *http.ServeMux
}

func NewApp(ctx context.Context, config *config.Config) (*App, error) {
	// postgres
	_, err := postgres.NewPostgresPool(config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// repo

	// service

	// mux register
	router := http.NewServeMux()

	return &App{
		Router: router,
	}, nil
}
