package main

import (
	"context"
	"fmt"
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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
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
	printInstructions(cfg.HTTPPort)

	// Ждём нажатия Ctrl+C.
	<-ctx.Done()

	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}

	slog.Info("server stopped")
}

func printInstructions(port string) {
	baseURL := "http://localhost:" + port

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("📚 BookLoop API")
	fmt.Println("=========================================")
	fmt.Println("Server started successfully!")
	fmt.Println()
	fmt.Println("Health:")
	fmt.Printf("  GET  %s/health\n", baseURL)
	fmt.Println()
	fmt.Println("Authentication:")
	fmt.Printf("  POST %s/api/v1/auth/register\n", baseURL)
	fmt.Printf("  POST %s/api/v1/auth/login\n", baseURL)
	fmt.Printf("  POST %s/api/v1/auth/refresh\n", baseURL)
	fmt.Printf("  POST %s/api/v1/auth/logout\n", baseURL)
	fmt.Println()
	fmt.Println("Books:")
	fmt.Printf("  GET  %s/api/v1/books\n", baseURL)
	fmt.Println("       Authorization: Bearer <access_token>")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server.")
	fmt.Println("=========================================")
	fmt.Println()
}
