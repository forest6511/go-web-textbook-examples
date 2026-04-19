package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery は panic を拾って構造化ログに記録し、500 で応答する
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, rec any) {
		logger.ErrorContext(c.Request.Context(),
			"panic recovered",
			"error", rec,
			"stack", string(debug.Stack()),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
	})
}
