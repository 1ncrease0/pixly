package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(addr, password string) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Redis{
		Client: rdb,
	}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
