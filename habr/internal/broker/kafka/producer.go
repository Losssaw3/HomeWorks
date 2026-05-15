package kafkawrap

import (
	"context"
	"encoding/json"
	"fmt"
	"sso/internal/domain/models"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:        kafka.TCP(brokers...),
			Topic:       topic,
			Balancer:    &kafka.LeastBytes{},
			MaxAttempts: 3,
		},
	}
}

func (p *Producer) Send(ctx context.Context, email string, uuid string) error {
	const op = "broker.Kafka.send"

	msg := models.Notification{
		Email:    email,
		UserUuid: uuid,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(email),
		Value: data,
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
