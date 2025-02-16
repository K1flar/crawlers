package create_task

import (
	"context"
	"log/slog"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/storage"
)

type Story struct {
	log   *slog.Logger
	tasks storage.Tasks
}

func NewStory(
	log *slog.Logger,
	tasks storage.Tasks,
) *Story {
	return &Story{log, tasks}
}

func (s *Story) Create(ctx context.Context, query string) (int64, error) {
	if query == "" {
		return 0, business_errors.InvalidQuery
	}

	id, err := s.tasks.Create(ctx, storage.ToCreateTask{
		Query: query,
	})

	return id, err
}
