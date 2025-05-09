package process_task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_model "github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/services"
	"github.com/K1flar/crawlers/internal/storage"
)

type Story struct {
	log                *slog.Logger
	tasksStorage       storage.Tasks
	taskSourcesStorage storage.TaskSources
	launcher           services.Launcher
	crawler            services.Crawler
	now                func() time.Time
}

func NewStory(
	log *slog.Logger,
	tasksStorage storage.Tasks,
	taskSourcesStorage storage.TaskSources,
	launcher services.Launcher,
	crawler services.Crawler,
) *Story {
	return &Story{
		log:                log,
		tasksStorage:       tasksStorage,
		taskSourcesStorage: taskSourcesStorage,
		launcher:           launcher,
		crawler:            crawler,
		now:                time.Now,
	}
}

func (s *Story) Process(ctx context.Context, id int64) error {
	s.log.Info(fmt.Sprintf("start to process task [%d]", id))

	task, err := s.tasksStorage.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if task.Status != task_model.StatusCreated && task.Status != task_model.StatusActive {
		return fmt.Errorf("task [%d] is not in created or active status (%s)", task.ID, task.Status)
	}

	err = s.tasksStorage.Process(ctx, id)
	if err != nil {
		return err
	}

	launchID, err := s.launcher.Start(ctx, id)
	if err != nil {
		return err
	}
	s.log.Info(fmt.Sprintf("new launch with id [%d]", launchID))

	pages, crawlerErr := s.crawler.Start(ctx, task)

	newStatus := task_model.StatusActive
	if crawlerErr != nil {
		newStatus = task_model.StatusStoppedWithError
	}

	err = s.launcher.Finish(ctx, services.LaunhToFinishParams{
		LaunchID: launchID,
		Task:     task,
		Pages:    pages,
		Error:    crawlerErr,
	})

	err = s.tasksStorage.SetStatus(ctx, id, newStatus)
	if err != nil {
		return err
	}

	return err
}
