package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	notifyapp "sso/internal/app/notification"
	notifyconfig "sso/internal/config/notification"
	"syscall"
)

const (
	envLocal  = "local"
	envDev    = "dev"
	envProd   = "prod"
	envDocker = "docker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := notifyconfig.MustLoad()
	logger := setupLogger(cfg.Env)

	application := notifyapp.NewNotificationApp(logger,
		cfg.StoragePath,
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
	)

	defer func() {
		logger.Info("performing cleanup...")
		application.Stop()
		logger.Info("gracefully stopped")
	}()

	if err := application.Run(ctx); err != nil {
		logger.Error("application stopped with error", "err", err)
	}
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
