package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/bookloop-alif/internal/delivery/http/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func signToken(t *testing.T, secret []byte, uid float64, role string) string {
	t.Helper()
	claims := jwt.MapClaims{"uid": uid, "role": role, "exp": 9999999999}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

func TestAuth_RejectsMissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	engine.Use(middleware.Auth([]byte("secret")))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = req
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_RejectsInvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(middleware.Auth([]byte("secret")))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_AcceptsValidToken(t *testing.T) {
	secret := []byte("secret")
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(middleware.Auth(secret))
	engine.GET("/", func(c *gin.Context) {
		if middleware.UserID(c) != 42 {
			t.Errorf("want uid=42, got %d", middleware.UserID(c))
		}
		if middleware.UserRole(c) != "user" {
			t.Errorf("want role=user, got %s", middleware.UserRole(c))
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, secret, 42, "user"))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestRequireRole_RejectsWrongRole(t *testing.T) {
	secret := []byte("secret")
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(middleware.Auth(secret), middleware.RequireRole("librarian"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, secret, 1, "user"))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestRequireRole_AllowsCorrectRole(t *testing.T) {
	secret := []byte("secret")
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(middleware.Auth(secret), middleware.RequireRole("librarian"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, secret, 1, "librarian"))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
