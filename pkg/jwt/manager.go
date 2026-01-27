package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrEmptySigningKey         = errors.New("empty signing key")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrExpiredToken            = errors.New("token expired")
	ErrClaimsNotFound          = errors.New("claims not found")
	ErrUserIDNotFound          = errors.New("user id not found in claims")
	ErrGenerateRefreshToken    = errors.New("failed to generate refresh token")
)

type Manager struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessTTL time.Duration, refreshTTL time.Duration) Manager {
	return Manager{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type AccessClaims struct {
	UserID uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

func (m *Manager) NewJWT(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

type Info struct {
	UserID uuid.UUID
}

func (m *Manager) Parse(accessToken string) (*Info, error) {
	token, err := jwt.ParseWithClaims(accessToken, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok {
		return nil, ErrClaimsNotFound
	}

	if claims.UserID == uuid.Nil {
		return nil, ErrUserIDNotFound
	}

	return &Info{
		UserID: claims.UserID,
	}, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ErrGenerateRefreshToken
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func VerifyRefreshToken(token, hashedToken string) bool {
	return HashRefreshToken(token) == hashedToken
}
