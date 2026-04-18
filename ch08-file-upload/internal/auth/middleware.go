package auth

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/apperror"
)

// JWTAuth は Authorization: Bearer <jwt> を検証するミドルウェア
func JWTAuth(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			_ = c.Error(apperror.NewUnauthorized(
				"missing bearer token", nil))
			c.Abort()
			return
		}
		claims, err := ParseAccessToken(raw, cfg)
		if err != nil {
			_ = c.Error(toAuthError(err))
			c.Abort()
			return
		}
		p := Principal{UserID: claims.UserID, Role: claims.Role}
		c.Set("principal", p)
		c.Request = c.Request.WithContext(
			WithPrincipal(c.Request.Context(), p))
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	return strings.TrimPrefix(header, prefix), true
}

func toAuthError(err error) *apperror.AppError {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return apperror.NewUnauthorized("token expired", err)
	case errors.Is(err, jwt.ErrTokenSignatureInvalid),
		errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenNotValidYet):
		return apperror.NewUnauthorized("invalid token", err)
	}
	return apperror.NewUnauthorized("authentication failed", err)
}
