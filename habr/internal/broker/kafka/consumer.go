package kafkawrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	log    *slog.Logger
}

func NewConsumer(log *slog.Logger, brokers []string, topic string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return &Consumer{
		reader: reader,
		log:    log,
	}

}

func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, error) {
	const op = "broker.Consumer.Fetch"
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("%s: %w", op, err)
	}

	return msg, nil
}

func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	const op = "broker.Consumer.Commit"
	err := c.reader.CommitMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Consumer) Close() error {
	const op = "broker.kafka.Consumer.Close"
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
