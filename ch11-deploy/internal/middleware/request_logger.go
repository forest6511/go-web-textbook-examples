package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type loggerKey struct{}

func RequestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Writer.Header().Set("X-Request-ID", reqID)
		logger := base.With(
			slog.String("request_id", reqID),
			slog.String("method", c.Request.Method),
			slog.String("route", c.FullPath()),
		)
		ctx := context.WithValue(c.Request.Context(),
			loggerKey{}, logger)
		c.Request = c.Request.WithContext(ctx)
		start := time.Now()
		c.Next()
		logger.LogAttrs(c.Request.Context(), slog.LevelInfo,
			"request completed",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.Int("size_bytes", c.Writer.Size()),
		)
	}
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
