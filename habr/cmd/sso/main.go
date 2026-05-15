package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	ssoapp "sso/internal/app/sso"
	config "sso/internal/config/sso"
	"syscall"
)

const (
	envLocal  = "local"
	envDev    = "dev"
	envProd   = "prod"
	envDocker = "docker"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)

	application := ssoapp.NewApp(logger, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL, cfg.Kafka.Brokers, cfg.Kafka.Topic)

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.GRPCServer.Run()
	}()

	select {
	case err := <-errCh:
		logger.Error("grpc server crashed", slog.Any("err", err))
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	}

	application.GRPCServer.Stop()
	logger.Info("Gracefully stopped")

	application.Storage.Stop()
	logger.Info("db connection closed")
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
	case envDocker:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
