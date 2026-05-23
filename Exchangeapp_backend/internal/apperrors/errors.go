package apperrors

import (
	"net/http"
)

type AppError struct {
	HTTPStatus int
	Code      int
	Message   string
}

func (appErr *AppError) Error() string {
	return appErr.Message
}

func New(status int, code int, message string) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message}
}

var (
	ErrInvalidParams = New(http.StatusBadRequest, 40001, "invalid params")
	ErrUnauthorized = New(http.StatusUnauthorized, 40101, "unauthorized")
	ErrForbidden = New(http.StatusForbidden, 40301, "forbidden")
	ErrNotFound = New(http.StatusNotFound, 40401, "not found")
	ErrInternal = New(http.StatusInternalServerError, 50001, "internal server error")
)