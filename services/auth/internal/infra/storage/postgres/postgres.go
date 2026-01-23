package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(conn string) (*Postgres, error) {
	pool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, fmt.Errorf("postgres.NewPostgres: %w", err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("postgres.NewPostgres: %w", err)
	}

	return &Postgres{
		Pool: pool,
	}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
