package activate_task

import (
	"log/slog"
	"net/http"

	"github.com/K1flar/crawlers/internal/handlers/common"
	"github.com/K1flar/crawlers/internal/message_broker"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
)

type Handler struct {
	log      *slog.Logger
	tasks    storage.Tasks
	producer message_broker.Producer[messages.TaskToProcessMessage]
}

func New(
	log *slog.Logger,
	tasks storage.Tasks,
	producer message_broker.Producer[messages.TaskToProcessMessage],
) *Handler {
	return &Handler{log, tasks, producer}
}

type dtoRequest struct {
	ID int64 `json:"id"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	defer func() {
		if err != nil {
			h.log.Error(err.Error())
		}
	}()

	dto, err := common.DTO[dtoRequest](r)
	if err != nil {
		common.BadRequest(w, "bad request body")
		return
	}

	err = h.tasks.SetStatus(ctx, dto.ID, task.StatusActive)
	if err != nil {
		common.Error(w, err)
		return
	}

	err = h.producer.Produce(ctx, messages.TaskToProcessMessage{ID: dto.ID})
	if err != nil {
		common.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
