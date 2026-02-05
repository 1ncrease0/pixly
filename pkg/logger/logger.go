package logger

import (
	"log/slog"
	"os"
)

const (
	Local = "local"
	Debug = "debug"
)

func Setup(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case Local:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case Debug:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
