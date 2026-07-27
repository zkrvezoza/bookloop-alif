package config_test

import (
	"testing"

	"github.com/bookloop-alif/internal/config"
)

func TestLoad_RequiresJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestLoad_UsesDefaultsWhenEnvNotSet(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBHost != "localhost" {
		t.Fatalf("want default localhost, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Fatalf("want default 5432, got %s", cfg.DBPort)
	}
}

func TestLoad_UsesProvidedEnv(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DB_HOST", "myhost")
	t.Setenv("DB_PORT", "5555")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBHost != "myhost" || cfg.DBPort != "5555" {
		t.Fatalf("env override didn't apply: %+v", cfg)
	}
}

func TestDSN_Format(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DB_HOST", "myhost")
	t.Setenv("DB_PORT", "5555")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "postgres://u:p@myhost:5555/db?sslmode=disable"
	if cfg.DSN() != want {
		t.Fatalf("want %s, got %s", want, cfg.DSN())
	}
}
