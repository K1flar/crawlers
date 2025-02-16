package storage

import "context"

type Tasks interface {
	Create(ctx context.Context, params ToCreateTask) (int64, error)
}

type Sources interface{}
