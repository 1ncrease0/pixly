package domain

import "errors"

var (
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUsernameTaken           = errors.New("username taken")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidUsername         = errors.New("invalid username")

	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExists   = errors.New("already exists")
)
