package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/1ncrease0/pixly/pkg/grpcserver"
	"github.com/1ncrease0/pixly/pkg/logger"
	"github.com/1ncrease0/pixly/services/art/internal/config"
	grpcart "github.com/1ncrease0/pixly/services/art/internal/infra/grpc/art"
	"github.com/1ncrease0/pixly/services/art/internal/infra/storage/minio"
	"github.com/1ncrease0/pixly/services/art/internal/infra/storage/postgres"
	"github.com/1ncrease0/pixly/services/art/internal/service/art"
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

	ctx := context.Background()
	m, err := minio.New(ctx, cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, cfg.MinIO.UseSSL, cfg.MinIO.Buckets)
	if err != nil {
		log.Error("failed to connect minio", slog.Any("error", err))
		return
	}

	bucket := cfg.MinIO.Buckets[0]
	imageProvider := minio.NewImageProvider(m.Client, bucket, cfg.MinIO.PresignTTL)

	pixelartRepo := postgres.NewPixelartRepo(pg.Pool)
	artService := art.NewService(pixelartRepo, &imageProvider, log)

	gRPCServer := grpcserver.New(log, cfg.GRPC.Port)
	grpcart.Register(gRPCServer.Server(), artService, log)

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
