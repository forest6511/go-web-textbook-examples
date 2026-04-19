package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/domain"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/repository"
)

type AuthUsecase struct {
	userRepo    *repository.UserRepo
	refreshRepo *repository.RefreshTokenRepo
	cfg         *auth.Config
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthUsecase(
	userRepo *repository.UserRepo,
	refreshRepo *repository.RefreshTokenRepo,
	cfg *auth.Config,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		cfg:         cfg,
	}
}

func (u *AuthUsecase) Signup(
	ctx context.Context, email, password string,
) (*TokenPair, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user, err := u.userRepo.Create(ctx, email, hash, string(auth.RoleUser))
	if err != nil {
		return nil, err
	}
	return u.issueTokens(ctx, user)
}

func (u *AuthUsecase) Login(
	ctx context.Context, email, password string,
) (*TokenPair, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			auth.ConsumeDummy(password)
			return nil, auth.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, err
	}
	return u.issueTokens(ctx, user)
}

func (u *AuthUsecase) Refresh(
	ctx context.Context, rawToken string,
) (*TokenPair, error) {
	oldHash := auth.HashRefreshToken(rawToken)
	newRaw, newHash, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	result, err := u.refreshRepo.Rotate(
		ctx, oldHash, newHash, time.Now().Add(u.cfg.RefreshTTL))
	if err != nil {
		return nil, err
	}
	user, err := u.userRepo.FindByID(ctx, result.UserID)
	if err != nil {
		return nil, err
	}
	access, err := auth.NewAccessToken(user.ID, user.Role, u.cfg)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: newRaw}, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, rawToken string) error {
	return u.refreshRepo.Revoke(ctx, auth.HashRefreshToken(rawToken))
}

func (u *AuthUsecase) issueTokens(
	ctx context.Context, user *domain.User,
) (*TokenPair, error) {
	access, err := auth.NewAccessToken(user.ID, user.Role, u.cfg)
	if err != nil {
		return nil, err
	}
	raw, hash, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	_, err = u.refreshRepo.InsertRoot(ctx, &repository.RefreshTokenRow{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(u.cfg.RefreshTTL),
	})
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: raw}, nil
}
