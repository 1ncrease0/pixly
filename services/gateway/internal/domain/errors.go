package domain

import "errors"

var (
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUsernameTaken           = errors.New("username taken")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserNotVerified         = errors.New("user not verified")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidUsername         = errors.New("invalid username")
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionExpired          = errors.New("session expired")
	ErrUnauthenticated         = errors.New("unauthenticated")

	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExists   = errors.New("already exists")
)
