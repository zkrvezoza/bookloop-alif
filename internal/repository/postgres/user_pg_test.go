package postgres_test

import (
	"context"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestUserRepo_CreateAndGetByLogin(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewUserRepo(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.User{
		Login: "khurliman", PasswordHash: "hash", Role: model.RoleUser,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	got, err := repo.GetByLogin(ctx, "khurliman")
	if err != nil {
		t.Fatalf("get by login failed: %v", err)
	}
	if got.ID != created.ID || got.Role != model.RoleUser {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestUserRepo_Create_DuplicateLogin(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewUserRepo(pool)
	ctx := context.Background()

	_, err := repo.Create(ctx, &model.User{Login: "khurliman", PasswordHash: "hash1", Role: model.RoleUser})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = repo.Create(ctx, &model.User{Login: "khurliman", PasswordHash: "hash2", Role: model.RoleUser})
	if err != model.ErrInvalid {
		t.Fatalf("want ErrInvalid on duplicate login, got %v", err)
	}
}

func TestUserRepo_GetByLogin_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewUserRepo(pool)

	_, err := repo.GetByLogin(context.Background(), "unknown")
	if err != model.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewUserRepo(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.User{Login: "ezoza", PasswordHash: "hash", Role: model.RoleLibrarian})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}
	if got.Login != "ezoza" || got.Role != model.RoleLibrarian {
		t.Fatalf("unexpected user: %+v", got)
	}
}
