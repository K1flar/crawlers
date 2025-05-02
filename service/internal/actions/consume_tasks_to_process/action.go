package consume_tasks_to_process

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/K1flar/crawlers/internal/message_broker"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/stories"
)

type Action struct {
	log              *slog.Logger
	consumer         message_broker.Consumer[messages.TaskToProcessMessage]
	story            stories.ProcessTask
	consumeBatchSize int
}

func NewAction(
	log *slog.Logger,
	consumer message_broker.Consumer[messages.TaskToProcessMessage],
	story stories.ProcessTask,
	consumeBatchSize int,
) *Action {
	return &Action{
		log:              log,
		consumer:         consumer,
		story:            story,
		consumeBatchSize: consumeBatchSize,
	}
}

func (a *Action) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for range a.consumeBatchSize {
		msg, err := a.consumer.Consume(ctx)
		if err != nil {
			a.log.Error(fmt.Sprintf("failed to consume task to process: %s", err.Error()))

			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.story.Process(ctx, msg.ID)
			if err != nil {
				a.log.Error(fmt.Sprintf("failed to consume task to process: %s", err.Error()))
			}
		}()
	}

	wg.Wait()
}
