package update_task

import (
	"log/slog"
	"net/http"

	"github.com/K1flar/crawlers/internal/handlers/common"
	"github.com/K1flar/crawlers/internal/storage"
)

type Handler struct {
	log   *slog.Logger
	tasks storage.Tasks
}

func New(
	log *slog.Logger,
	tasks storage.Tasks,
) *Handler {
	return &Handler{log, tasks}
}

type dtoRequest struct {
	ID                     int64    `json:"id"`
	DepthLevel             *int     `json:"depthLevel"`
	MinWeight              *float64 `json:"minWeight"`
	MaxSources             *int64   `json:"maxSources"`
	MaxNeighboursForSource *int64   `json:"maxNeighboursForSource"`
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

	err = h.tasks.Update(ctx, storage.ToUpdateTask{
		ID:                     dto.ID,
		DepthLevel:             dto.DepthLevel,
		MinWeight:              dto.MinWeight,
		MaxSources:             dto.MaxSources,
		MaxNeighboursForSource: dto.MaxNeighboursForSource,
	})
	if err != nil {
		common.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
