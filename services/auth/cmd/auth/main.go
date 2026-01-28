package main

import (
	"context"
	"github.com/1ncrease0/pixly/pkg/grpcserver"
	"github.com/1ncrease0/pixly/pkg/jwt"
	"github.com/1ncrease0/pixly/pkg/logger"
	"github.com/1ncrease0/pixly/services/auth/internal/config"
	grpcauth "github.com/1ncrease0/pixly/services/auth/internal/infra/grpc/auth"
	"github.com/1ncrease0/pixly/services/auth/internal/infra/rabbitmq"
	"github.com/1ncrease0/pixly/services/auth/internal/infra/storage/postgres"
	"github.com/1ncrease0/pixly/services/auth/internal/infra/storage/redis"
	"github.com/1ncrease0/pixly/services/auth/internal/service/auth"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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
		log.Error("failed to connect rabbitmq", slog.Any("error", err))
		return
	}
	defer func() {
		err = producer.Close()
		if err != nil {
			slog.Any("error", err)
		}
	}()

	gRPCServer := grpcserver.New(log, cfg.GRPC.Port)

	verificationRepo := redis.NewVerificationRepo(rds.Client, cfg.Verification.TTL)
	userRepo := postgres.NewUserRepo(pg.Pool)
	sessionRepo := postgres.NewSessionRepo(pg.Pool)
	jwtManger := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	authService := auth.NewAuthService(userRepo, verificationRepo, producer, sessionRepo, jwtManger)
	grpcauth.Register(gRPCServer.Server(), authService, log)

	gRPCServer.Start()
	defer gRPCServer.Shutdown()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-gRPCServer.Notify():
		log.Error("grpc server error", slog.Any("error", err))
	case <-ctx.Done():
		log.Info("shutting down")
	}
}
