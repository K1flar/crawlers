package services

import (
	"context"

	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/K1flar/crawlers/internal/models/task"
)

type Crawler interface {
	Start(ctx context.Context, task task.Task) (map[string]*page.PageWithParentURL, error)
}

type CollectionCollector interface {
	AddPage(url string, page page.Page)
	BM25(url string) (float64, bool)
}

type Launcher interface {
	Start(ctx context.Context, taskID int64) (int64, error)
	Finish(ctx context.Context, params LaunhToFinishParams) error
}
