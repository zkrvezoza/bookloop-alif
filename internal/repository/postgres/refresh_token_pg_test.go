package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestRefreshTokenRepo_CreateAndGetByHash(t *testing.T) {
	pool := setupTestDB(t)
	userID, _ := seedUserAndBook(t, pool, 1)
	repo := postgres.NewRefreshTokenRepo(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.RefreshToken{
		UserID: userID, TokenHash: "hash123", ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := repo.GetByHash(ctx, "hash123")
	if err != nil {
		t.Fatalf("get by hash failed: %v", err)
	}
	if got.ID != created.ID || got.UserID != userID {
		t.Fatalf("unexpected token: %+v", got)
	}
}

func TestRefreshTokenRepo_GetByHash_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewRefreshTokenRepo(pool)

	_, err := repo.GetByHash(context.Background(), "unknown")
	if err != model.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRefreshTokenRepo_Revoke(t *testing.T) {
	pool := setupTestDB(t)
	userID, _ := seedUserAndBook(t, pool, 1)
	repo := postgres.NewRefreshTokenRepo(pool)
	ctx := context.Background()

	rt, _ := repo.Create(ctx, &model.RefreshToken{UserID: userID, TokenHash: "h", ExpiresAt: time.Now().Add(time.Hour)})

	if err := repo.Revoke(ctx, rt.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByHash(ctx, "h")
	if err != nil {
		t.Fatalf("get after revoke failed: %v", err)
	}
	if got.RevokedAt == nil {
		t.Fatal("expected revoked_at to be set")
	}
}
