package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{
		db: pool,
	}
}

func (r *SessionRepo) CreateSession(ctx context.Context, session *domain.Session) error {
	if session == nil {
		return domain.ErrNilSession
	}

	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, session.UserID, session.TokenHash, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepo) Session(ctx context.Context, tokenHash string) (*domain.Session, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var (
		id        int64
		userID    uuid.UUID
		hash      string
		expiresAt time.Time
		createdAt time.Time
	)

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&id,
		&userID,
		&hash,
		&expiresAt,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	if time.Now().After(expiresAt) {
		_ = r.DeleteSession(ctx, tokenHash)
		return nil, domain.ErrSessionExpired
	}

	return &domain.Session{
		ID:        id,
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}, nil
}

func (r *SessionRepo) DeleteSession(ctx context.Context, tokenHash string) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

func (r *SessionRepo) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}
