package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bookloop-alif/internal/config"
	deliveryhttp "github.com/bookloop-alif/internal/delivery/http"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DSN())
	if err != nil {
		slog.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	engine := deliveryhttp.New(deliveryhttp.Deps{
		Pool:      pool,
		JWTSecret: cfg.JWTSecret,
	})
	srv := deliveryhttp.NewServer(engine, cfg.HTTPPort)

	if err := srv.Start(); err != nil {
		slog.Error("server start failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server started", "port", cfg.HTTPPort)

	<-ctx.Done()
	slog.Info("shutting down...")

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}

	slog.Info("stopped")
}
