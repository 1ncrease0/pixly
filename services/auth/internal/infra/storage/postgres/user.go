package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: pool,
	}
}

func (r *UserRepo) UserByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	const query = `
		SELECT id, email, name, password_hash, is_verified
		FROM users
		WHERE email = $1
	`

	var (
		id           uuid.UUID
		rawEmail     string
		rawName      string
		passwordHash string
		isVerified   bool
	)

	err := r.db.QueryRow(ctx, query, email.String()).Scan(
		&id,
		&rawEmail,
		&rawName,
		&passwordHash,
		&isVerified,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	emailVO, err := domain.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}

	nameVO, err := domain.NewUsername(rawName)
	if err != nil {
		return nil, err
	}

	user := domain.NewUser(
		id,
		emailVO,
		nameVO,
		domain.NewPasswordFromHash(passwordHash),
		isVerified,
	)

	return user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrNilUser
	}

	const query = `
        INSERT INTO users (id, name, email, password_hash, is_verified)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (email) 
        DO UPDATE SET
			id = EXCLUDED.id,
            name = EXCLUDED.name,
            password_hash = EXCLUDED.password_hash
        WHERE users.is_verified = false
        RETURNING id
    `

	var returnedID string
	err := r.db.QueryRow(ctx, query,
		user.ID().String(),
		user.Name().String(),
		user.Email().String(),
		user.PasswordHash(),
		user.IsVerified(),
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserAlreadyExists
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUsernameTaken
		}

		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepo) UserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, email, name, password_hash, is_verified
		FROM users
		WHERE id = $1
	`

	var (
		id           uuid.UUID
		rawEmail     string
		rawName      string
		passwordHash string
		isVerified   bool
	)

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&id,
		&rawEmail,
		&rawName,
		&passwordHash,
		&isVerified,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	emailVO, err := domain.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}

	nameVO, err := domain.NewUsername(rawName)
	if err != nil {
		return nil, err
	}

	user := domain.NewUser(
		id,
		emailVO,
		nameVO,
		domain.NewPasswordFromHash(passwordHash),
		isVerified,
	)

	return user, nil
}

func (r *UserRepo) Verify(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE users
		SET is_verified = true
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
