package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	articleApp "sso/internal/app/article"
	articleconfig "sso/internal/config/article"
	"syscall"
	"time"
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

	cfg := articleconfig.MustLoad()
	logger := setupLogger(cfg.Env)

	application := articleApp.NewArticleApp(logger,
		cfg.Name,
		cfg.Secret,
		cfg.StoragePath,
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDBNum,
		cfg.RedisTTL,
		cfg.HttpAddr,
		cfg.GRPCAddr,
		2*time.Second,
	)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- application.Srv.Run()
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	case err := <-srvErr:
		logger.Error("http server crashed", slog.Any("err", err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	application.Stop(shutdownCtx)
	logger.Info("application stopped")
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
