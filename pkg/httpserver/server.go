package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Params struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
	log             *slog.Logger
	notify          chan error
}

func New(p Params, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         p.Addr,
			Handler:      handler,
			ReadTimeout:  p.ReadTimeout,
			WriteTimeout: p.WriteTimeout,
			IdleTimeout:  p.IdleTimeout,
		},
		shutdownTimeout: p.ShutdownTimeout,
		log:             logger,
		notify:          make(chan error, 1),
	}
}

func (s *Server) Start() {
	go func() {
		s.log.Info("http server started", slog.String("addr", s.httpServer.Addr))

		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("http server failed", slog.Any("error", err))
			s.notify <- err
		}
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	s.log.Info("http server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("http server forced shutdown", slog.Any("error", err))
		_ = s.httpServer.Close()
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
