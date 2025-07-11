package app

import (
	"log/slog"
	grpcapp "main/internal/app/grpc"
	"main/internal/config"
)

type App struct {
	GrpcServer *grpcapp.App
}

func NewApp(log *slog.Logger, grpcPort int, dbConfig config.DbConfig) *App {
	//TODO: Init service
	//TODO: Init db

	grpcApp := grpcapp.NewApp(log, grpcPort)

	return &App{
		GrpcServer: grpcApp,
	}
}
