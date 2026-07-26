package service

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

const (
	accessTTL      = 15 * time.Minute
	refreshTTL     = 7 * 24 * time.Hour
	minPasswordLen = 6
)

//US-03

type AuthService struct {
	users     repository.UserRepo
	tokens    repository.RefreshTokenRepo
	jwtSecret []byte
}

func NewAuthService(users repository.UserRepo, tokens repository.RefreshTokenRepo, jwtSecret string) *AuthService {
	return &AuthService{users: users, tokens: tokens, jwtSecret: []byte(jwtSecret)}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Register — US-03. Роль выбирается при регистрации: user или librarian.
func (s *AuthService) Register(ctx context.Context, login, password string, role model.Role) (*model.User, error) {
	if login == "" || len(password) < minPasswordLen {
		return nil, model.ErrInvalid
	}
	if role != model.RoleUser && role != model.RoleLibrarian {
		return nil, model.ErrInvalid
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.users.Create(ctx, &model.User{Login: login, PasswordHash: string(hash), Role: role})
}
