package main

import (
	"context"
	"github.com/1ncrease0/pixly/pkg/httpserver"
	"github.com/1ncrease0/pixly/pkg/jwt"
	"github.com/1ncrease0/pixly/pkg/logger"
	"github.com/1ncrease0/pixly/services/gateway/internal/config"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api"
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
	}
	defer func() {
		err = authClient.Close()
		if err != nil {
			log.Error("failed to close auth client", slog.Any("error", err))
		}
	}()

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	router := api.InitRoutes(cfg, log, authClient, jwtManager)
	httpServer := httpserver.New(
		httpserver.Params{
			Addr:            cfg.HTTP.Host + ":" + strconv.Itoa(cfg.HTTP.Port),
			ReadTimeout:     cfg.HTTP.ReadTimeout,
			WriteTimeout:    cfg.HTTP.WriteTimeout,
			ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
		},
		router,
		log,
	)

	httpServer.Start()
	defer func() {
		err = httpServer.Shutdown()
		if err != nil {
			log.Error("failed to shutdown http server", slog.Any("error", err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-httpServer.Notify():
		log.Error("grpc server error", slog.Any("error", err))
	case <-ctx.Done():
		log.Info("shutting down")
	}

}
