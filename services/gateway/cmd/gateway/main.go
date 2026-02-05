package main

import (
	"context"
	"github.com/1ncrease0/pixly/pkg/httpserver"
	"github.com/1ncrease0/pixly/pkg/jwt"
	"github.com/1ncrease0/pixly/pkg/logger"
	"github.com/1ncrease0/pixly/services/gateway/internal/config"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/grpc/art"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/grpc/auth"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

	authClient, err := auth.NewClient(cfg.Clients.Auth.Addr, cfg.Clients.Auth.Timeout, cfg.Clients.Auth.Retries, log)
	if err != nil {
		log.Error("failed to create auth client", slog.Any("error", err))
		return
	}
	defer func() {
		err = authClient.Close()
		if err != nil {
			log.Error("failed to close auth client", slog.Any("error", err))
		}
	}()

	artClient, err := art.NewClient(cfg.Clients.Art.Addr, cfg.Clients.Art.Timeout, cfg.Clients.Art.Retries, log)
	if err != nil {
		log.Error("failed to create art client", slog.Any("error", err))
		return
	}
	defer func() {
		err = artClient.Close()
		if err != nil {
			log.Error("failed to close art client", slog.Any("error", err))
		}
	}()

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	router := api.InitRoutes(cfg, log, authClient, artClient, jwtManager)
	httpServer := httpserver.New(
		httpserver.Params{
			Addr:            cfg.HTTP.Host + ":" + strconv.Itoa(cfg.HTTP.Port),
			ReadTimeout:     cfg.HTTP.ReadTimeout,
			WriteTimeout:    cfg.HTTP.WriteTimeout,
			IdleTimeout:     cfg.HTTP.IdleTimeout,
			ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
		},
		router,
		log,
	)

	httpServer.Start()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-httpServer.Notify():
		log.Error("http server error", slog.Any("error", err))
	case <-ctx.Done():
		log.Info("shutting down")
	}

	if err := httpServer.Shutdown(); err != nil {
		log.Error("failed to shutdown http server", slog.Any("error", err))
	}
}
