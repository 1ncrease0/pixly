package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrNilSession      = errors.New("nil session")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type Session struct {
	ID        int64
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type TokenPair struct {
	Access  string
	Refresh string
}
