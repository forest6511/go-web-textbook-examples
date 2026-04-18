package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders は OWASP 推奨ヘッダを一括付与する
// production=true のときだけ HSTS を送る（ローカル開発を壊さないため）
func SecurityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy",
			"default-src 'self'; frame-ancestors 'none'; object-src 'none'")
		h.Set("Permissions-Policy",
			"camera=(), microphone=(), geolocation=()")

		if production {
			h.Set("Strict-Transport-Security",
				"max-age=63072000; includeSubDomains")
		}

		c.Next()
	}
}
