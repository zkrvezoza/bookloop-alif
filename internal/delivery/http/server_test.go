package http

import (
	"context"
	stdhttp "net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	server := NewServer(engine, "9999")

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	if server.httpServer == nil {
		t.Fatal("expected non-nil http server")
	}

	if server.httpServer.Addr != ":9999" {
		t.Fatalf(
			"expected address %q, got %q",
			":9999",
			server.httpServer.Addr,
		)
	}

	if server.httpServer.Handler != engine {
		t.Fatal("expected Gin engine to be the server handler")
	}

	if server.httpServer.ReadTimeout != 10*time.Second {
		t.Fatalf(
			"expected read timeout %v, got %v",
			10*time.Second,
			server.httpServer.ReadTimeout,
		)
	}

	if server.httpServer.WriteTimeout != 10*time.Second {
		t.Fatalf(
			"expected write timeout %v, got %v",
			10*time.Second,
			server.httpServer.WriteTimeout,
		)
	}
}

func TestServer_Shutdown_NotStarted(t *testing.T) {
	engine := gin.New()
	server := NewServer(engine, "0")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil && err != stdhttp.ErrServerClosed {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
}
