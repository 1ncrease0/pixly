package main

import (
	"context"
	"github.com/1ncrease0/pixly/pkg/logger"
	"github.com/1ncrease0/pixly/services/notification/internal/config"
	"github.com/1ncrease0/pixly/services/notification/internal/infra/email"
	"github.com/1ncrease0/pixly/services/notification/internal/infra/rabbitmq"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

	sender := email.NewSender(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.Email.From,
		cfg.Email.Base,
	)

	consumer, err := rabbitmq.NewConsumer(
		log,
		sender,
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.Queue,
		cfg.RabbitMQ.Workers,
	)
	if err != nil {
		log.Error("consumer init failed", slog.Any("error", err))
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := consumer.Start(ctx); err != nil {
		log.Error("consumer start failed", slog.Any("error", err))
		return
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = consumer.Shutdown(shutdownCtx)

}
