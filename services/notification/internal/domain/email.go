package domain

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email")
)

type Email struct {
	email string
}

func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Email{}, ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(raw); err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{
		email: raw,
	}, nil
}

func (e Email) String() string {
	return e.email
}
