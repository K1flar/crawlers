package process_task

import (
	"context"
	"fmt"
	"log/slog"

	task_model "github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/services"
	"github.com/K1flar/crawlers/internal/storage"
)

type Story struct {
	log     *slog.Logger
	storage storage.Tasks
	crawler services.Crawler
}

func NewStory(
	log *slog.Logger,
	storage storage.Tasks,
	crawler services.Crawler,
) *Story {
	return &Story{
		log:     log,
		storage: storage,
		crawler: crawler,
	}
}

func (s *Story) Process(ctx context.Context, id int64) error {
	s.log.Info(fmt.Sprintf("start to process task [%d]", id))

	task, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if task.Status != task_model.StatusActive {
		return fmt.Errorf("task [%d] is not in active status (%s)", task.ID, task.Status)
	}

	return s.crawler.Start(ctx, task)
}
