package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/domain/model"
)

func JSON(c *gin.Context, code int, data any) {
	c.JSON(code, data)
}

// Error маппит sentinel-ошибки из model в HTTP-коды.
// Тело ошибки всегда {"error": "..."}.
func Error(c *gin.Context, err error) {
	code := http.StatusInternalServerError

	switch {
	case errors.Is(err, model.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, model.ErrInvalid):
		code = http.StatusBadRequest
	case errors.Is(err, model.ErrForbidden):
		code = http.StatusForbidden
	case errors.Is(err, model.ErrNoSeats):
		code = http.StatusConflict
	}

	c.JSON(code, gin.H{"error": err.Error()})
}
