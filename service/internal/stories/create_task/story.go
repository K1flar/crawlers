package create_task

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/message_broker"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/storage"
)

const (
	maxCountWords = 10
	maxLenWord    = 20

	defaultDepthLevel             = 3
	defaultMinWeight              = 0
	defaultMaxSources             = 20
	defaultMaxNeighboursForSource = 20
)

type Story struct {
	log      *slog.Logger
	tasks    storage.Tasks
	producer message_broker.Producer[messages.TaskToProcessMessage]
}

func NewStory(
	log *slog.Logger,
	tasks storage.Tasks,
	producer message_broker.Producer[messages.TaskToProcessMessage],
) *Story {
	return &Story{log, tasks, producer}
}

func (s *Story) Create(ctx context.Context, query string) (int64, error) {
	if err := s.validateQuery(query); err != nil {
		return 0, err
	}

	id, err := s.tasks.Create(ctx, storage.ToCreateTask{
		Query:                  query,
		DepthLevel:             defaultDepthLevel,
		MinWeight:              defaultMinWeight,
		MaxSources:             defaultMaxSources,
		MaxNeighboursForSource: defaultMaxNeighboursForSource,
	})
	if err != nil {
		return 0, err
	}

	err = s.producer.Produce(ctx, messages.TaskToProcessMessage{
		ID: id,
	})
	if err != nil {
		return 0, err
	}

	s.log.Info(fmt.Sprintf("create and produce task [%d]: [%s]", id, query))

	return id, err
}

func (s *Story) validateQuery(query string) error {
	if query == "" {
		return business_errors.InvalidQuery
	}

	words := strings.Split(query, " ")

	if len(words) > maxCountWords {
		return business_errors.InvalidQuery
	}

	for _, w := range words {
		if len([]rune(w)) > maxLenWord {
			return business_errors.InvalidQuery
		}
	}

	return nil
}
