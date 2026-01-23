package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidPassword = errors.New("invalid password")
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

type Username struct {
	username string
}

func NewUsername(raw string) (Username, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Username{}, ErrInvalidUsername
	}
	return Username{raw}, nil
}

func (u Username) String() string {
	return u.username
}

type Password struct {
	hash string
}

func NewPassword(raw string) (Password, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 8 {
		return Password{}, ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, err
	}

	return Password{hash: string(hash)}, nil
}

func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

func (p Password) Hash() string {
	return p.hash
}

func (p Password) Equals(raw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(raw))
	return err == nil
}

type VerificationCode struct {
	value string
}

func NewVerificationCodeFromString(raw string) VerificationCode {
	raw = strings.TrimSpace(raw)
	return VerificationCode{value: raw}
}

func NewVerificationCode() VerificationCode {
	return VerificationCode{value: uuid.NewString()}
}

func (c VerificationCode) String() string {
	return c.value
}

func (c VerificationCode) Hash() string {
	hash := sha256.Sum256([]byte(c.value))
	return hex.EncodeToString(hash[:])
}
