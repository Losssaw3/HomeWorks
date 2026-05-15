package notifyapp

import (
	"context"
	"errors"
	"log/slog"
	kafkawrap "sso/internal/broker/kafka"
	"sso/internal/mail"
	"sso/internal/services/notification"
	"sso/internal/storage/postgres"
	"time"
)

type NotificationApp struct {
	Handler *notification.NotificationHandler
	Storage *postgres.Storage
}

func NewNotificationApp(log *slog.Logger,
	pathToStorage string,
	brokers []string,
	topic string,
	groupID string,

) *NotificationApp {
	storage, err := postgres.NewStorage(pathToStorage)
	if err != nil {
		panic(err)
	}

	consumer := kafkawrap.NewConsumer(log, brokers, topic, groupID)
	mailer := &mail.Mailer{}

	handler := notification.NewNotificationHandler(log, consumer, mailer, storage)

	return &NotificationApp{Handler: handler, Storage: storage}
}

func (a *NotificationApp) Run(ctx context.Context) error {
	const op = "app.notification.run"

	for {
		select {
		case <-ctx.Done():
			a.Handler.Log.Info("stopping notification app", slog.String("operation", op))
			return ctx.Err()
		default:
			if err := a.Handler.ProcessVerification(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				a.Handler.Log.Error("failed to process verification", "err", err)
			}
			time.Sleep(time.Second)
			continue
		}
	}

}

func (a *NotificationApp) Stop() {
	const op = "app.notification.stop"

	a.Handler.Log.With(slog.String("op", op)).Info("stopping notification app")

	a.Storage.Stop()
	a.Handler.Consumer.Close()
}
