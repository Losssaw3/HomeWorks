package ssoapp

import (
	"log/slog"
	grpcapp "sso/internal/app/sso/grpc"
	kafkawrap "sso/internal/broker/kafka"
	"sso/internal/services/auth"
	"time"

	"sso/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	Storage    *postgres.Storage
}

func NewApp(log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
	kafkaBrokers []string,
	kafkaTopic string,
) *App {
	storage, err := postgres.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}
	producer := kafkawrap.NewProducer(kafkaBrokers, kafkaTopic)
	authService := auth.NewAuth(log, storage, storage, storage, producer, tokenTTL)

	grpcApp := grpcapp.NewApp(log, *authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
		Storage:    storage,
	}

}
