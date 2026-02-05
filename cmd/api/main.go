package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"web_demoservice/internal/app"
	"web_demoservice/internal/config"
)

func main() {
	cfg, err := config.NewConfig("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port), Handler: *api.Router}

	go func() {
		if err = server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	slog.Info("Server started on ", slog.Any("addr", server.Addr), slog.Any("port", cfg.HTTP.Port))
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if err = server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
