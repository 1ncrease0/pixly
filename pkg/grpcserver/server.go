package grpcserver

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
)

type Server struct {
	gRPCServer *grpc.Server
	log        *slog.Logger
	port       int
	notify     chan error
}

func New(log *slog.Logger, port int) *Server {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("recovered from panic", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
		))

	return &Server{
		gRPCServer: gRPCServer,
		log:        log,
		port:       port,
		notify:     make(chan error, 1),
	}
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (s *Server) Server() *grpc.Server {
	return s.gRPCServer
}

func (s *Server) Start() {
	go func() {
		defer close(s.notify)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
		if err != nil {
			s.log.Error("failed to listen", slog.Any("error", err))
			s.notify <- fmt.Errorf("failed to listen: %w", err)

			return
		}

		s.log.Info("grpc server started", slog.String("addr", l.Addr().String()))
		if err := s.gRPCServer.Serve(l); err != nil {
			s.log.Error("failed to serve", slog.Any("error", err))
			s.notify <- err
		}
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() {
	s.log.Info("grpc server shutting down")
	s.gRPCServer.GracefulStop()
}
