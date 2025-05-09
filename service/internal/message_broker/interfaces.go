package message_broker

import "context"

type Producer[T any] interface {
	Produce(ctx context.Context, message T) error
}

type Consumer[T any] interface {
	Consume(ctx context.Context) (T, error)
}
