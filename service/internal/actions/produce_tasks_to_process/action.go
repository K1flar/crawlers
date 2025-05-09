package produce_tasks_to_process

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/K1flar/crawlers/internal/stories"
)

type Action struct {
	log   *slog.Logger
	story stories.ProduceTasksToProcess
}

func NewAction(
	log *slog.Logger,
	story stories.ProduceTasksToProcess,
) *Action {
	return &Action{
		log:   log,
		story: story,
	}
}

func (a *Action) Run(ctx context.Context) {
	err := a.story.ProduceAll(ctx)
	if err != nil {
		a.log.Error(fmt.Sprintf("failed to produce active tasks: %s", err.Error()))
	}
}
