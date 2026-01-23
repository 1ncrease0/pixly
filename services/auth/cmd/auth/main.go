package main

import (
	"context"
	"github.com/1ncrease0/pixly/pkg/logger"
	"github.com/1ncrease0/pixly/services/auth/internal/config"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/1ncrease0/pixly/services/auth/internal/infra/rabbitmq"
	"github.com/1ncrease0/pixly/services/auth/internal/infra/storage/postgres"
	"github.com/1ncrease0/pixly/services/auth/internal/infra/storage/redis"
	"github.com/1ncrease0/pixly/services/auth/internal/service/auth"
	"log/slog"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

	pg, err := postgres.New(cfg.Postgres.DSN)
	if err != nil {
		log.Error("failed to connect postgres", slog.Any("error", err))
		return
	}
	defer pg.Close()

	rds, err := redis.New(cfg.Redis.Addr, cfg.Redis.Password)
	if err != nil {
		log.Error("failed to connect redis", slog.Any("error", err))
		return
	}
	defer func() {
		err = rds.Close()
		if err != nil {
			slog.Any("error", err)
		}
	}()

	producer, err := rabbitmq.NewProducer(cfg.RabbitMQ.URL, cfg.RabbitMQ.Queue)
	if err != nil {
		log.Error("failed to connect kafka", slog.Any("error", err))
		return
	}
	defer func() {
		err = producer.Close()
		if err != nil {
			slog.Any("error", err)
		}
	}()

	verificationRepo := redis.NewVerificationRepo(rds.Client, cfg.Verification.TTL)
	userRepo := postgres.NewUserRepo(pg.Pool)

	authService := auth.NewAuthService(userRepo, verificationRepo, producer)

	email, _ := domain.NewEmail("email@gmail.com")
	password, _ := domain.NewPassword("12345678")
	name, _ := domain.NewUsername("username")
	err = authService.Register(context.Background(), email, name, password)
	if err != nil {
		log.Error("failed to register auth", slog.Any("error", err))
	}

	time.Sleep(10 * time.Second)

}
