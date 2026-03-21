package errors

import (
	"fmt"
	"net/http"
)

// AppError is the application-level error type that carries an error code,
// a human-readable message, and an HTTP status code for response mapping.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Inner      error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Inner)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the inner error for errors.Is and errors.As compatibility.
func (e *AppError) Unwrap() error {
	return e.Inner
}

// NotFound returns an AppError indicating a resource was not found.
func NotFound(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// BadRequest returns an AppError indicating a malformed or invalid request.
func BadRequest(message string) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// Unauthorized returns an AppError indicating missing or invalid authentication.
func Unauthorized() *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "authentication required",
		StatusCode: http.StatusUnauthorized,
	}
}

// Forbidden returns an AppError indicating the user lacks permission.
func Forbidden() *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    "insufficient permissions",
		StatusCode: http.StatusForbidden,
	}
}

// Internal returns an AppError for unexpected server-side failures.
// The actual error is wrapped and logged but not exposed to the client.
func Internal(err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "an internal error occurred",
		StatusCode: http.StatusInternalServerError,
		Inner:      err,
	}
}

// ValidationFailed returns an AppError indicating request validation failure.
// It accepts a map of field names to error messages for structured feedback.
func ValidationFailed(message string) *AppError {
	return &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

// Conflict returns an AppError indicating a duplicate resource conflict.
func Conflict(message string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// RateLimited returns an AppError indicating the client has exceeded rate limits.
func RateLimited() *AppError {
	return &AppError{
		Code:       "RATE_LIMITED",
		Message:    "too many requests, please try again later",
		StatusCode: http.StatusTooManyRequests,
	}
}

// PaymentRequired returns an AppError for billing/quota limits.
func PaymentRequired(message string) *AppError {
	return &AppError{
		Code:       "PAYMENT_REQUIRED",
		Message:    message,
		StatusCode: http.StatusPaymentRequired,
	}
}

// RequireResult checks that a repository result is non-nil.
// It returns a NOT_FOUND AppError when val is nil and err is nil.
// Any other error is wrapped with Internal.
func RequireResult[T any](val *T, err error, resource string) (*T, error) {
	if err != nil {
		return nil, Internal(err)
	}
	if val == nil {
		return nil, NotFound(resource)
	}
	return val, nil
}
