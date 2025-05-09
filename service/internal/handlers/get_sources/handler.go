package get_sources

import (
	"log/slog"
	"net/http"

	"github.com/K1flar/crawlers/internal/handlers/common"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/samber/lo"
)

type Handler struct {
	log     *slog.Logger
	sources storage.Sources
}

func New(
	log *slog.Logger,
	sources storage.Sources,
) *Handler {
	return &Handler{log, sources}
}

type dtoRequest struct {
	ID int64 `json:"id"`
}

type dtoResponse struct {
	Sources []dtoSource `json:"sources"`
}

type dtoSource struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	Weight   float64 `json:"weight"`
	ParentID *int64  `json:"parentId"`
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

	sources, err := h.sources.GetByTaskID(ctx, dto.ID)
	if err != nil {
		common.Error(w, err)
		return
	}

	common.OK(w, dtoResponse{
		Sources: lo.Map(sources, func(source source.ForTask, _ int) dtoSource {
			return dtoSource{
				ID:       source.ID,
				URL:      source.URL,
				Title:    source.Title,
				Weight:   source.Weight,
				ParentID: source.ParentID,
			}
		}),
	})
}
