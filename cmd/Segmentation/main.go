package main

import (
	"context"
	"github.com/joho/godotenv"
	"log/slog"
	"main/internal/app"
	"main/internal/config"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("failed to load .env file: " + err.Error())
	}

	cfg := config.MustLoadConfig()

	log := setupLogger(cfg.Env)

	application := app.NewApp(log, cfg.Grpc.Port, cfg.Db, cfg.Cache, cfg.Queue)

	go application.GrpcServer.MustRun()
	go application.KafkaConsumer.MustRun(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop

	log.Info("Stopping app with signal " + sig.String())

	application.GrpcServer.Stop()

	log.Info("gracefully stopped grpc server")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
