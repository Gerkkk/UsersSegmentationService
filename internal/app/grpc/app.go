package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	segmentationrpc "main/internal/grpc/segmentation"
	"net"
)

// App - структура Grpc сервера, который использует приложение
type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

// NewApp - конструктор App
func NewApp(log *slog.Logger, port int, segmentationService segmentationrpc.Segmentation) *App {
	grpcServer := grpc.NewServer()

	segmentationrpc.Register(grpcServer, segmentationService)

	return &App{
		log:        log,
		grpcServer: grpcServer,
		port:       port,
	}
}

// MustRun - Запуск Grpc сервера. При ошибке паникует
func (a *App) MustRun() {
	const op = "grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("Starting grpc server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		panic("Grpc server failed to listen: " + err.Error())
	}

	log.Info("GRPC server listening on", slog.String("address", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		panic("Grpc server failed to serve: " + err.Error())
	}
}

// Stop - Graceful stop grpc-сервера
func (a *App) Stop() error {
	const op = "grpcapp.Stop"

	log := a.log.With(slog.String("op", op))
	log.Info("stopping grpc server", slog.Int("port", a.port))

	a.grpcServer.GracefulStop()

	return nil
}
