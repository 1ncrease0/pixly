package apierr

import (
	"errors"
	"github.com/1ncrease0/pixly/services/gateway/internal/domain"
	"net/http"
)

type ErrorResponse struct {
	Error  ErrorDetail `json:"error"`
	Status int
}

func NewErrorResponse(status int, code Code, message, field string) ErrorResponse {
	return ErrorResponse{
		Status: status,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Field:   field,
		},
	}
}

type ErrorDetail struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type Code string

const (
	BadRequest              Code = "BAD_REQUEST"
	InvalidEmail            Code = "INVALID_EMAIL"
	InvalidUsername         Code = "INVALID_USERNAME"
	InvalidPassword         Code = "INVALID_PASSWORD"
	InvalidVerificationCode Code = "INVALID_VERIFICATION_CODE"
	UserAlreadyExists       Code = "USER_ALREADY_EXISTS"
	UsernameTaken           Code = "USERNAME_TAKEN"
	UserNotFound            Code = "USER_NOT_FOUND"
	InternalError           Code = "INTERNAL_ERROR"
)

func ErrorToResponse(err error) ErrorResponse {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail):
		return NewErrorResponse(http.StatusBadRequest, InvalidEmail, "Invalid email format", "email")
	case errors.Is(err, domain.ErrInvalidUsername):
		return NewErrorResponse(http.StatusBadRequest, InvalidUsername, "Invalid username", "username")
	case errors.Is(err, domain.ErrInvalidPassword):
		return NewErrorResponse(http.StatusBadRequest, InvalidPassword, "Password must be at least 8 characters", "password")
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return NewErrorResponse(http.StatusConflict, UserAlreadyExists, "User with this email already exists", "email")
	case errors.Is(err, domain.ErrUsernameTaken):
		return NewErrorResponse(http.StatusConflict, UsernameTaken, "Username is already taken", "username")
	case errors.Is(err, domain.ErrInvalidVerificationCode):
		return NewErrorResponse(http.StatusBadRequest, InvalidVerificationCode, "Invalid or expired verification code", "code")
	case errors.Is(err, domain.ErrUserNotFound):
		return NewErrorResponse(http.StatusNotFound, UserNotFound, "User not found", "")
	default:
		return NewErrorResponse(http.StatusInternalServerError, InternalError, "Internal server error", "")
	}
}
