package create_task

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/handlers/common"
	"github.com/K1flar/crawlers/internal/stories"
)

type Handler struct {
	log   *slog.Logger
	story stories.CreateTask
}

func New(
	log *slog.Logger,
	story stories.CreateTask,
) *Handler {
	return &Handler{log, story}
}

type dtoRequest struct {
	Query string `json:"query"`
}

type dtoResponse struct {
	ID int64 `json:"id"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	dto, err := common.DTO[dtoRequest](r)
	if err != nil {
		common.BadRequest(w, "bad request body")
		return
	}

	id, err := h.story.Create(r.Context(), dto.Query)
	if err != nil {
		var businessError *business_errors.BusinessError
		if errors.As(err, &businessError) {
			common.Forbidden(w, businessError.Code, common.ErrorMsg(err))
			return
		}

		common.InternalError(w)
		return
	}

	common.OK(w, dtoResponse{id})
}
