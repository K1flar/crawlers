package produce_tasks_to_process

import (
	"context"

	"github.com/K1flar/crawlers/internal/message_broker"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
)

type Story struct {
	storage  storage.Tasks
	producer message_broker.Producer[messages.TaskToProcessMessage]
}

func NewStory(
	storage storage.Tasks,
	producer message_broker.Producer[messages.TaskToProcessMessage],
) *Story {
	return &Story{
		storage:  storage,
		producer: producer,
	}
}

func (s *Story) ProduceAll(ctx context.Context) error {
	tasks, err := s.storage.FindInStatuses(ctx, []task.Status{task.StatusActive})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err := s.producer.Produce(ctx, messages.TaskToProcessMessage{
			ID: task.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
