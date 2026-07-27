package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.JSON(c, http.StatusCreated, gin.H{"ok": true})

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestError_MapsSentinelErrorsToStatusCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", model.ErrNotFound, http.StatusNotFound},
		{"invalid", model.ErrInvalid, http.StatusBadRequest},
		{"forbidden", model.ErrForbidden, http.StatusForbidden},
		{"no seats", model.ErrNoSeats, http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			response.Error(c, tt.err)

			if w.Code != tt.want {
				t.Fatalf("want %d, got %d", tt.want, w.Code)
			}

			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid json body: %v", err)
			}
			if body["error"] == "" {
				t.Fatal("expected error message in body")
			}
		})
	}
}
