package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	segmentationrpc "main/internal/grpc/segmentation"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

func NewApp(log *slog.Logger, port int, segmentationService segmentationrpc.Segmentation) *App {
	grpcServer := grpc.NewServer()

	segmentationrpc.Register(grpcServer, segmentationService)

	return &App{
		log:        log,
		grpcServer: grpcServer,
		port:       port,
	}
}

func (a *App) MustRun() error {
	const op = "grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("Starting grpc server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("GRPC server listening on", slog.String("address", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() error {
	const op = "grpcapp.Stop"

	log := a.log.With(slog.String("op", op))
	log.Info("stopping grpc server", slog.Int("port", a.port))

	a.grpcServer.GracefulStop()

	return nil
}
