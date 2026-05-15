package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sso/internal/broker/kafka"
	"sso/internal/domain/models"
)

type EmailSender interface {
	SendConfirmation(ctx context.Context, email string, body string) error
}

type UserConfirmer interface {
	Confirm(ctx context.Context, email string) error
}

type NotificationHandler struct {
	Log       *slog.Logger
	Consumer  *kafkawrap.Consumer
	Mailer    EmailSender
	Confirmer UserConfirmer
}

func NewNotificationHandler(log *slog.Logger,
	consumer *kafkawrap.Consumer,
	mailer EmailSender,
	confirmer UserConfirmer,
) *NotificationHandler {
	return &NotificationHandler{
		Log:       log,
		Consumer:  consumer,
		Mailer:    mailer,
		Confirmer: confirmer,
	}
}

func (h *NotificationHandler) ProcessVerification(ctx context.Context) error {
	const op = "services.Notification.ProcessVerification"
	msg, err := h.Consumer.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var payload models.Notification
	err = json.Unmarshal(msg.Value, &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log := h.Log.With(
		slog.String("op", op),
		slog.String("email", payload.Email),
	)
	log.Info("confirming user email")

	err = h.Mailer.SendConfirmation(ctx, payload.Email, payload.UserUuid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	//todo добавить ожидание ответа от подтверждения
	err = h.Confirmer.Confirm(ctx, payload.Email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = h.Consumer.Commit(ctx, msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
