package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type VerificationRepo struct {
	client  *redis.Client
	codeTTL time.Duration
}

func NewVerificationRepo(client *redis.Client, codeTTL time.Duration) *VerificationRepo {
	return &VerificationRepo{
		client:  client,
		codeTTL: codeTTL,
	}
}

func (r *VerificationRepo) key(codeHash string) string {
	return fmt.Sprintf("verification:%s", codeHash)
}

func (r *VerificationRepo) Save(ctx context.Context, codeHash string, userID uuid.UUID) error {
	return r.client.Set(ctx, r.key(codeHash), userID.String(), r.codeTTL).Err()
}

func (r *VerificationRepo) Get(ctx context.Context, codeHash string) (uuid.UUID, error) {
	val, err := r.client.Get(ctx, r.key(codeHash)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, domain.ErrVerificationCodeNotFound
		}
		return uuid.Nil, err
	}
	return uuid.Parse(val)
}

func (r *VerificationRepo) Delete(ctx context.Context, codeHash string) error {
	return r.client.Del(ctx, r.key(codeHash)).Err()
}
