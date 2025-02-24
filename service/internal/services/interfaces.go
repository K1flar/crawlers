package services

import (
	"context"

	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/K1flar/crawlers/internal/models/task"
)

type Crawler interface {
	Start(ctx context.Context, task task.Task) error
}

type CollectionCollector interface {
	AddPage(uuid string, page page.Page)
	BM25(uuid string) (float64, bool)
}
