package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
func (s *AuthService) Login(ctx context.Context, login, password string) (*TokenPair, error) {
	u, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrInvalid
		}
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, model.ErrInvalid
	}

	return s.issueTokenPair(ctx, u)
}

//US-04 Логин + refresh-токены

func (s *AuthService) Refresh(ctx context.Context, rawRefresh string) (*TokenPair, error) {
	rt, err := s.tokens.GetByHash(ctx, hashToken(rawRefresh))
	if err != nil {
		return nil, model.ErrInvalid
	}
	if rt.RevokedAt != nil || rt.ExpiresAt.Before(time.Now()) {
		return nil, model.ErrInvalid
	}

	u, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.tokens.Revoke(ctx, rt.ID); err != nil {
		return nil, err
	}

	return s.issueTokenPair(ctx, u)
}

func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return model.ErrInvalid
	}

	rt, err := s.tokens.GetByHash(ctx, hashToken(rawRefresh))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil
		}
		return err
	}

	if rt.RevokedAt != nil {
		return nil
	}

	return s.tokens.Revoke(ctx, rt.ID)
}

func (s *AuthService) issueTokenPair(ctx context.Context, u *model.User) (*TokenPair, error) {
	access, err := s.signAccessToken(u)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(refreshTTL),
	}
	if _, err := s.tokens.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: access, RefreshToken: rawRefresh}, nil
}

func (s *AuthService) signAccessToken(u *model.User) (string, error) {
	claims := jwt.MapClaims{
		"uid":  u.ID,
		"role": string(u.Role),
		"exp":  time.Now().Add(accessTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
