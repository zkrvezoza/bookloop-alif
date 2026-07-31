package http

import (
	stdhttp "net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNew_RegistersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := New(Deps{
		Pool:      nil,
		JWTSecret: "test-secret-key",
	})

	routes := engine.Routes()

	expected := map[string]bool{
		stdhttp.MethodGet + " /health":                false,
		stdhttp.MethodPost + " /api/v1/auth/register": false,
		stdhttp.MethodPost + " /api/v1/auth/login":    false,
		stdhttp.MethodPost + " /api/v1/auth/refresh":  false,
		stdhttp.MethodPost + " /api/v1/auth/logout":   false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path

		if _, exists := expected[key]; exists {
			expected[key] = true
		}
	}

	for route, found := range expected {
		if !found {
			t.Errorf("expected route %s to be registered", route)
		}
	}
}

func TestNew_ReturnsEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := New(Deps{
		Pool:      nil,
		JWTSecret: "test-secret-key",
	})

	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
}
