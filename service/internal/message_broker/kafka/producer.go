package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer[T any] struct {
	writer *kafka.Writer
	now    func() time.Time
}

func NewProducer[T any](brokers []string, topic string) *Producer[T] {
	config := kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	return &Producer[T]{
		writer: kafka.NewWriter(config),
		now:    time.Now,
	}
}

func (p *Producer[T]) Produce(ctx context.Context, msg T) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: b,
		Time:  p.now(),
	})
}

func (p *Producer[T]) Close() error {
	return p.writer.Close()
}
