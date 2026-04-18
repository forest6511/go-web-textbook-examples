package domain

import "errors"

var (
	ErrTaskNotFound   = errors.New("task not found")
	ErrDuplicate      = errors.New("duplicate record")
	ErrForeignKey     = errors.New("foreign key violation")
	ErrCheckViolation = errors.New("check constraint violation")
)
