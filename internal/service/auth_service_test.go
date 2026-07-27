package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/service"
)

type fakeUserRepo struct {
	byLogin map[string]*model.User
	nextID  int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byLogin: make(map[string]*model.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, u *model.User) (*model.User, error) {
	if _, exists := f.byLogin[u.Login]; exists {
		return nil, model.ErrInvalid
	}
	f.nextID++
	u.ID = f.nextID
	f.byLogin[u.Login] = u
	return u, nil
}

func (f *fakeUserRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	u, ok := f.byLogin[login]
	if !ok {
		return nil, model.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	for _, u := range f.byLogin {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, model.ErrNotFound
}

type fakeRefreshTokenRepo struct {
	tokens map[string]*model.RefreshToken
	nextID int64
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{tokens: make(map[string]*model.RefreshToken)}
}

func (f *fakeRefreshTokenRepo) Create(_ context.Context, rt *model.RefreshToken) (*model.RefreshToken, error) {
	f.nextID++
	rt.ID = f.nextID
	f.tokens[rt.TokenHash] = rt
	return rt, nil
}

func (f *fakeRefreshTokenRepo) GetByHash(_ context.Context, hash string) (*model.RefreshToken, error) {
	rt, ok := f.tokens[hash]
	if !ok {
		return nil, model.ErrNotFound
	}
	return rt, nil
}

func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id int64) error {
	for _, rt := range f.tokens {
		if rt.ID == id {
			now := rt.ExpiresAt
			rt.RevokedAt = &now
			return nil
		}
	}
	return model.ErrNotFound
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		role     model.Role
		wantErr  error
	}{
		{"valid user", "khurliman", "password123", model.RoleUser, nil},
		{"valid librarian", "ezoza", "password123", model.RoleLibrarian, nil},
		{"empty login", "", "password123", model.RoleUser, model.ErrInvalid},
		{"short password", "khurliman", "123", model.RoleUser, model.ErrInvalid},
		{"invalid role", "khurliman", "password123", model.Role("admin"), model.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := newFakeUserRepo()
			tokens := newFakeRefreshTokenRepo()
			svc := service.NewAuthService(users, tokens, "test-secret")

			_, err := svc.Register(context.Background(), tt.login, tt.password, tt.role)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuthService_Register_DuplicateLogin(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	svc := service.NewAuthService(users, tokens, "test-secret")
	ctx := context.Background()

	_, err := svc.Register(ctx, "aziza", "password123", model.RoleUser)
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = svc.Register(ctx, "aziza", "password456", model.RoleUser)
	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("want ErrInvalid for duplicate login, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	svc := service.NewAuthService(users, tokens, "test-secret")
	ctx := context.Background()

	_, err := svc.Register(ctx, "aziza", "password123", model.RoleUser)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		pair, err := svc.Login(ctx, "aziza", "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pair.AccessToken == "" || pair.RefreshToken == "" {
			t.Fatal("expected non-empty tokens")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, "aziza", "wrongpass")
		if !errors.Is(err, model.ErrInvalid) {
			t.Fatalf("want ErrInvalid, got %v", err)
		}
	})

	t.Run("unknown login", func(t *testing.T) {
		_, err := svc.Login(ctx, "unknown", "password123")
		if !errors.Is(err, model.ErrInvalid) {
			t.Fatalf("want ErrInvalid, got %v", err)
		}
	})
}

func TestAuthService_RefreshAndLogout(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	svc := service.NewAuthService(users, tokens, "test-secret")
	ctx := context.Background()

	svc.Register(ctx, "khurliman", "password123", model.RoleUser)
	pair, _ := svc.Login(ctx, "khurliman", "password123")

	t.Run("refresh rotates token", func(t *testing.T) {
		newPair, err := svc.Refresh(ctx, pair.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newPair.RefreshToken == pair.RefreshToken {
			t.Fatal("expected refresh token to rotate")
		}

		_, err = svc.Refresh(ctx, pair.RefreshToken)
		if !errors.Is(err, model.ErrInvalid) {
			t.Fatalf("want ErrInvalid for revoked token, got %v", err)
		}

		pair = newPair
	})

	t.Run("logout revokes refresh", func(t *testing.T) {
		if err := svc.Logout(ctx, pair.RefreshToken); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.Refresh(ctx, pair.RefreshToken)
		if !errors.Is(err, model.ErrInvalid) {
			t.Fatalf("want ErrInvalid after logout, got %v", err)
		}
	})
}
