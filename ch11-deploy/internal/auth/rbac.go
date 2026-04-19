package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch11-deploy/internal/apperror"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// RequireRole は JWTAuth の後段で呼ばれ、role ミスマッチなら 403
func RequireRole(required Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := PrincipalFromContext(c.Request.Context())
		if !ok {
			_ = c.Error(apperror.NewForbidden(
				"principal missing", nil))
			c.Abort()
			return
		}
		if Role(p.Role) != required {
			_ = c.Error(apperror.NewForbidden(
				"role not allowed", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}
