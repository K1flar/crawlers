package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Consumer[T any] struct {
	reader *kafka.Reader
}

func NewConsumer[T any](brokers []string, topic string) *Consumer[T] {
	config := kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
	}

	return &Consumer[T]{
		reader: kafka.NewReader(config),
	}
}

func (c *Consumer[T]) Consume(ctx context.Context) (T, error) {
	var res T

	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return res, fmt.Errorf("failed to read message from kafka: %w", err)
	}

	if err := json.Unmarshal(msg.Value, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (c *Consumer[T]) Close() error {
	return c.reader.Close()
}
