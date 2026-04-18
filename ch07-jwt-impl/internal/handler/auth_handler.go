package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/apperror"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/usecase"
)

type AuthHandler struct {
	usecase *usecase.AuthUsecase
	cfg     *auth.Config
}

func NewAuthHandler(u *usecase.AuthUsecase, cfg *auth.Config) *AuthHandler {
	return &AuthHandler{usecase: u, cfg: cfg}
}

type SignupRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	tokens, err := h.usecase.Signup(
		c.Request.Context(), req.Email, req.Password)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	setRefreshCookie(c, tokens.RefreshToken, h.cfg)
	c.JSON(http.StatusCreated, gin.H{
		"access_token": tokens.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.cfg.AccessTTL.Seconds()),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	tokens, err := h.usecase.Login(
		c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			_ = c.Error(apperror.NewUnauthorized(
				"invalid credentials", err))
			return
		}
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	setRefreshCookie(c, tokens.RefreshToken, h.cfg)
	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.cfg.AccessTTL.Seconds()),
	})
}

func setRefreshCookie(c *gin.Context, token string, cfg *auth.Config) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token", token,
		int(cfg.RefreshTTL.Seconds()),
		"/api/v1/auth",
		"",
		true,
		true,
	)
}
