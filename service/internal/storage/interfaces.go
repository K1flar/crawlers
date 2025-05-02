package storage

import (
	"context"

	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/models/task"
)

type Tasks interface {
	GetByID(ctx context.Context, id int64) (task.Task, error)
	FindInStatuses(ctx context.Context, statuses []task.Status) ([]task.Task, error)
	Create(ctx context.Context, params ToCreateTask) (int64, error)
	SetStatus(ctx context.Context, id int64, status task.Status) error
}

type Sources interface {
	GetByTaskID(ctx context.Context, taskID int64) ([]source.Source, error)
	Create(ctx context.Context, params []ToCreateSource) ([]int64, error)
}
