package create_task

import (
	"context"
	"log/slog"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/services"
	"github.com/K1flar/crawlers/internal/storage"
)

type Story struct {
	log     *slog.Logger
	tasks   storage.Tasks
	crawler services.Crawler
}

func NewStory(
	log *slog.Logger,
	tasks storage.Tasks,
	crawler services.Crawler,
) *Story {
	return &Story{log, tasks, crawler}
}

func (s *Story) Create(ctx context.Context, query string) (int64, error) {
	if query == "" {
		return 0, business_errors.InvalidQuery
	}

	id, err := s.tasks.Create(ctx, storage.ToCreateTask{
		Query: query,
	})
	if err != nil {
		return 0, err
	}

	task, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}

	err = s.crawler.Start(ctx, task)

	return id, err
}
