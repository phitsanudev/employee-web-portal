package domain

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrValidation   = errors.New("validation error")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal error")
)

type AppError struct {
	Kind    error
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Kind
}

func NewAppError(kind error, code, message string) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message}
}
