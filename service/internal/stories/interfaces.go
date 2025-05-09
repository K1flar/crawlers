package stories

import (
	"context"
)

type CreateTask interface {
	Create(ctx context.Context, query string) (int64, error)
}

type ProduceTasksToProcess interface {
	ProduceAll(ctx context.Context) error
}

type ProcessTask interface {
	Process(ctx context.Context, id int64) error
}
