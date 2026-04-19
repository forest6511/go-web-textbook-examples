package apperror

import (
	"errors"
	"net/http"

	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/domain"
)

func NewValidation(
	msg string, details []FieldIssue, cause error,
) *AppError {
	return &AppError{
		Code:     "VALIDATION_FAILED",
		Message:  msg,
		HTTPCode: http.StatusUnprocessableEntity, // 422
		Cause:    cause,
		Details:  details,
	}
}

func NewBadRequest(msg string, cause error) *AppError {
	return &AppError{
		Code: "BAD_REQUEST", Message: msg,
		HTTPCode: http.StatusBadRequest, Cause: cause,
	}
}

func NewNotFound(msg string, cause error) *AppError {
	return &AppError{
		Code: "NOT_FOUND", Message: msg,
		HTTPCode: http.StatusNotFound, Cause: cause,
	}
}

func NewConflict(msg string, cause error) *AppError {
	return &AppError{
		Code: "CONFLICT", Message: msg,
		HTTPCode: http.StatusConflict, Cause: cause,
	}
}

func NewInternal(cause error) *AppError {
	return &AppError{
		Code: "INTERNAL", Message: "internal server error",
		HTTPCode: http.StatusInternalServerError, Cause: cause,
	}
}

func NewUnauthorized(msg string, cause error) *AppError {
	return &AppError{
		Code: "UNAUTHORIZED", Message: msg,
		HTTPCode: http.StatusUnauthorized, Cause: cause,
	}
}

func NewForbidden(msg string, cause error) *AppError {
	return &AppError{
		Code: "FORBIDDEN", Message: msg,
		HTTPCode: http.StatusForbidden, Cause: cause,
	}
}

func NewPayloadTooLarge(msg string, cause error) *AppError {
	return &AppError{
		Code: "PAYLOAD_TOO_LARGE", Message: msg,
		HTTPCode: http.StatusRequestEntityTooLarge, Cause: cause,
	}
}

func NewUnsupportedMediaType(msg string, cause error) *AppError {
	return &AppError{
		Code: "UNSUPPORTED_MEDIA_TYPE", Message: msg,
		HTTPCode: http.StatusUnsupportedMediaType, Cause: cause,
	}
}

// FromDomain はドメインエラーを HTTP 表現に変換する
func FromDomain(err error) *AppError {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		return NewNotFound("task not found", err)
	case errors.Is(err, domain.ErrDuplicate):
		return NewConflict("duplicate record", err)
	case errors.Is(err, domain.ErrCheckViolation):
		return NewBadRequest("invalid value", err)
	case errors.Is(err, domain.ErrForeignKey):
		return NewBadRequest("referenced resource missing", err)
	}
	return NewInternal(err)
}
