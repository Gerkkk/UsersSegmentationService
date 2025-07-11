package app

import (
	"log/slog"
	grpcapp "main/internal/app/grpc"
	"main/internal/config"
	"main/internal/repository/postgres"
	"main/internal/services/segmentation"
)

type App struct {
	GrpcServer *grpcapp.App
}

func NewApp(log *slog.Logger, grpcPort int, dbConfig config.DbConfig) *App {

	shards := make([]string, 0)

	for _, cfg := range dbConfig.Shards {
		shards = append(shards, cfg.DSN)
	}

	repository, err := postgres.NewSegmentationStorage(dbConfig.NumShards, shards)

	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	segService := segmentation.NewSegmentation(log, repository)

	grpcApp := grpcapp.NewApp(log, grpcPort, segService)

	return &App{
		GrpcServer: grpcApp,
	}
}
