package app

import (
	"log/slog"
	grpcapp "main/internal/app/grpc"
	"main/internal/app/kafka"
	"main/internal/config"
	kafkahandler "main/internal/kafka"
	"main/internal/repository/postgres"
	"main/internal/repository/redis"
	"main/internal/services/segmentation"
	"main/internal/services/users"
)

type App struct {
	GrpcServer    *grpcapp.App
	KafkaConsumer *kafka.App
}

func NewApp(log *slog.Logger, grpcPort int, dbConfig config.DbConfig, cacheConfig config.CacheConfig, queueConfig config.QueueConfig) *App {

	shards := make([]string, 0)

	for _, cfg := range dbConfig.Shards {
		shards = append(shards, cfg.DSN)
	}

	repository, err := postgres.NewSegmentationStorage(dbConfig.NumShards, shards)

	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	segCache, err := redis.NewSegmentationCache(log, cacheConfig.Host, cacheConfig.Port, cacheConfig.Password, cacheConfig.MaxMemory, cacheConfig.MaxMemoryPolicy)

	if err != nil {
		panic("failed to connect to cache: " + err.Error())
	}

	segService := segmentation.NewSegmentation(log, repository, segCache)

	grpcApp := grpcapp.NewApp(log, grpcPort, segService)

	userService := users.NewUsers(log, repository)

	messageHandler := kafkahandler.New(log, userService)

	kafkaApp := kafka.New(log, messageHandler, queueConfig.Brokers, queueConfig.Topics, queueConfig.Group)

	return &App{
		GrpcServer:    grpcApp,
		KafkaConsumer: kafkaApp,
	}
}
