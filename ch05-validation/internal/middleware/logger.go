package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger は slog でアクセスログを構造化出力する
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()
		logResponse(c, logger, path, raw, start)
	}
}

func logResponse(
	c *gin.Context, logger *slog.Logger,
	path, raw string, start time.Time,
) {
	latency := time.Since(start)
	status := c.Writer.Status()
	rid, _ := c.Get("request_id")
	if raw != "" {
		path = path + "?" + raw
	}
	logger.LogAttrs(c.Request.Context(),
		slog.LevelInfo, "http_request",
		slog.String("request_id", toString(rid)),
		slog.String("method", c.Request.Method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Duration("latency", latency),
		slog.String("client_ip", c.ClientIP()),
	)
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
