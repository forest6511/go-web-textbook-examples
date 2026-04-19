package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/apperror"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/domain"
)

func (h *AuthHandler) Refresh(c *gin.Context) {
	raw, err := c.Cookie("refresh_token")
	if err != nil || raw == "" {
		_ = c.Error(apperror.NewUnauthorized(
			"missing refresh token", err))
		return
	}
	tokens, err := h.usecase.Refresh(c.Request.Context(), raw)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRefreshReused):
			_ = c.Error(apperror.NewUnauthorized(
				"token reuse detected", err))
		case errors.Is(err, domain.ErrRefreshRevoked),
			errors.Is(err, domain.ErrRefreshExpired):
			_ = c.Error(apperror.NewUnauthorized(
				"refresh token invalid", err))
		default:
			_ = c.Error(apperror.FromDomain(err))
		}
		return
	}
	setRefreshCookie(c, tokens.RefreshToken, h.cfg)
	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.cfg.AccessTTL.Seconds()),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	raw, err := c.Cookie("refresh_token")
	if err != nil || raw == "" {
		c.Status(http.StatusNoContent)
		return
	}
	_ = h.usecase.Logout(c.Request.Context(), raw)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", "", -1,
		"/api/v1/auth", "", h.cfg.SecureCookie, true)
	c.Status(http.StatusNoContent)
}
