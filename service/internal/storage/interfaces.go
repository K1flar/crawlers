package storage

import (
	"context"

	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/models/task"
)

type Tasks interface {
	GetByID(ctx context.Context, id int64) (task.Task, error)
	GetForList(ctx context.Context, filter FilterTaskForList) ([]task.ForList, error)
	GetCount(ctx context.Context) (int64, error)
	FindInStatuses(ctx context.Context, statuses []task.Status) ([]task.Task, error)
	Create(ctx context.Context, params ToCreateTask) (int64, error)
	SetStatus(ctx context.Context, id int64, status task.Status) error
	Process(ctx context.Context, id int64) error
	Update(ctx context.Context, params ToUpdateTask) error
}

type Sources interface {
	Create(ctx context.Context, params []ToCreateSource) (map[string]int64, error)
	Update(ctx context.Context, params []ToUpdateSource) (map[string]int64, error)
	GetByURLs(ctx context.Context, urls []string) (map[string]source.Source, error)
	GetByTaskID(ctx context.Context, taskID int64) ([]source.ForTask, error)
	GetForProtocol(ctx context.Context, filter FilterForProtocol) ([]source.ForProtocol, error)
}

type TaskSources interface {
	Create(ctx context.Context, params []ToCreateTaskSource) error
}

type Launches interface {
	Create(ctx context.Context, params ToCreateLaunch) (int64, error)
	Finish(ctx context.Context, params ToFinishLaunch) error
	Get(ctx context.Context, id int64) (launch.Launch, error)
	GetLastByTaskID(ctx context.Context, taskID int64) (launch.Launch, error)
}
