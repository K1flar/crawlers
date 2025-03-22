package kafka

import (
	"context"
)

type MessageBroker[T any] struct {
	topic string
}

func NewMessageBroker[T any](topic string) *MessageBroker[T] {
	return &MessageBroker[T]{topic}
}

func (b *MessageBroker[T]) Produce(ctx context.Context, message T) error {
	return nil
}

func (b *MessageBroker[T]) Consume(ctx context.Context) (T, error) {
	var res T
	return res, nil
}
