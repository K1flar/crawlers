package storage

import (
	"context"

	"github.com/K1flar/crawlers/internal/models/source"
)

type Tasks interface {
	Create(ctx context.Context, params ToCreateTask) (int64, error)
}

type Sources interface {
	GetByTaskID(ctx context.Context, taskID int64) ([]source.Source, error)
	Create(ctx context.Context, params []ToCreateSource) error
}
