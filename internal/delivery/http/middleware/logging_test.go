package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/delivery/http/middleware"
)

func TestRequestID_SetsHeaderAndContext(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(middleware.RequestID())

	var gotID string
	engine.GET("/", func(c *gin.Context) {
		gotID = middleware.RequestIDFromCtx(c.Request.Context())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
	if gotID == "" {
		t.Fatal("expected request id to be available via context")
	}
}

func TestRequestIDFromCtx_EmptyWhenNotSet(t *testing.T) {
	if got := middleware.RequestIDFromCtx(context.Background()); got != "" {
		t.Fatalf("want empty string, got %q", got)
	}
}

func TestAccessLog_DoesNotBreakRequest(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(middleware.RequestID(), middleware.AccessLog())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Fatalf("want 418, got %d", w.Code)
	}
}
