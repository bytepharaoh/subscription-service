package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

var (
	ErrNotFound      = &AppError{Code: "NOT_FOUND", Message: "resource not found", StatusCode: http.StatusNotFound}
	ErrInvalidInput  = &AppError{Code: "INVALID_INPUT", Message: "invalid input", StatusCode: http.StatusBadRequest}
	ErrInvalidDate   = &AppError{Code: "INVALID_DATE", Message: "invalid date format, use MM-YYYY", StatusCode: http.StatusBadRequest}
	ErrInvalidPeriod = &AppError{Code: "INVALID_PERIOD", Message: "period_start must be before period_end", StatusCode: http.StatusBadRequest}
	ErrInternal      = &AppError{Code: "INTERNAL_ERROR", Message: "internal server error", StatusCode: http.StatusInternalServerError}
)

func New(base *AppError, detail string) *AppError {
	return &AppError{
		Code:       base.Code,
		Message:    fmt.Sprintf("%s: %s", base.Message, detail),
		StatusCode: base.StatusCode,
	}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
