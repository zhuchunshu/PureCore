package core

import "net/http"

// ValidationError represents a validation error with a map of field-level errors
type ValidationError struct {
	Message string
	Errors  map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// AppError represents an application-level error with an HTTP status code
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// NewAppErrorf creates a new application error with a formatted message
func NewAppErrorf(code int, format string, args ...interface{}) *AppError {
	msg := format
	if len(args) > 0 {
		msg = http.StatusText(code)
	}
	return &AppError{Code: code, Message: msg}
}

// Common application errors
var (
	ErrNotFound         = NewAppError(404, "Resource not found")
	ErrUnauthorized     = NewAppError(401, "Unauthorized")
	ErrForbidden        = NewAppError(403, "Forbidden")
	ErrInternalServer   = NewAppError(500, "Internal server error")
	ErrBadRequest       = NewAppError(400, "Bad request")
	ErrMethodNotAllowed = NewAppError(405, "Method not allowed")
	ErrTooManyRequests  = NewAppError(429, "Too many requests")
)
