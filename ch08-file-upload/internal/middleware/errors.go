package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/apperror"
)

// Errors は c.Errors を RFC 7807 形式で書き出す
func Errors(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		appErr := toAppError(err)
		if appErr.HTTPCode >= 500 {
			logger.ErrorContext(c.Request.Context(),
				"request failed", "error", err, "code", appErr.Code)
		}
		c.Header("Content-Type", "application/problem+json")
		c.AbortWithStatusJSON(
			appErr.HTTPCode,
			apperror.ToProblem(appErr, c.GetString("request_id")),
		)
	}
}

func toAppError(err error) *apperror.AppError {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return apperror.NewValidation(
			"request body did not pass validation",
			toFieldIssues(ve), err,
		)
	}
	var syn *json.SyntaxError
	var ute *json.UnmarshalTypeError
	if errors.As(err, &syn) || errors.As(err, &ute) {
		return apperror.NewBadRequest(
			"request body is not valid JSON", err)
	}
	return apperror.NewInternal(err)
}

func toFieldIssues(ve validator.ValidationErrors) []apperror.FieldIssue {
	out := make([]apperror.FieldIssue, 0, len(ve))
	for _, fe := range ve {
		out = append(out, apperror.FieldIssue{
			Field:   fe.Field(), // JSON タグ名マッピング済み
			Tag:     fe.Tag(),
			Param:   fe.Param(),
			Message: fe.Error(),
		})
	}
	return out
}
