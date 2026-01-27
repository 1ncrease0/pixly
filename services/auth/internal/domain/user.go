package domain

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrUsernameTaken            = errors.New("user username taken")
	ErrNilUser                  = errors.New("nil user")
	ErrVerificationCodeNotFound = errors.New("verification not found")
	ErrUserAlreadyVerified      = errors.New("user already verified")
	ErrUserNotVerified          = errors.New("user not verified")
)

type User struct {
	id         uuid.UUID
	email      Email
	name       Username
	password   Password
	isVerified bool
}

func NewUser(id uuid.UUID, email Email, name Username, password Password, isVerified bool) *User {
	return &User{
		id:         id,
		email:      email,
		name:       name,
		password:   password,
		isVerified: isVerified,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}
func (u *User) Email() Email {
	return u.email
}
func (u *User) Name() Username {
	return u.name
}
func (u *User) PasswordHash() string {
	return u.password.Hash()
}
func (u *User) IsVerified() bool {
	return u.isVerified
}
