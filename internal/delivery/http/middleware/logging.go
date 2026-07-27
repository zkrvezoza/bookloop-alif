package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()
		c.Set(string(requestIDKey), id)

		ctx := context.WithValue(c.Request.Context(), requestIDKey, id)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID, _ := c.Get(string(requestIDKey))
		slog.Info("http request",
			"request_id", reqID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func RequestIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
