package domain

import "errors"

var (
	ErrTaskNotFound   = errors.New("task not found")
	ErrDuplicate      = errors.New("duplicate record")
	ErrForeignKey     = errors.New("foreign key violation")
	ErrCheckViolation = errors.New("check constraint violation")

	ErrUserNotFound   = errors.New("user not found")
	ErrRefreshRevoked = errors.New("refresh token revoked")
	ErrRefreshReused  = errors.New("refresh token reused")
	ErrRefreshExpired = errors.New("refresh token expired")
)
