package get_task

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/K1flar/crawlers/internal/handlers/common"
	task_model "github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/internal/utils"
)

type Handler struct {
	log      *slog.Logger
	tasks    storage.Tasks
	launches storage.Launches
}

func New(
	log *slog.Logger,
	tasks storage.Tasks,
	launches storage.Launches,
) *Handler {
	return &Handler{log, tasks, launches}
}

type dtoRequest struct {
	ID int64 `json:"id"`
}

type dtoResponse struct {
	Query                  string         `json:"query"`
	Status                 string         `json:"status"`
	CreatedAt              time.Time      `json:"createdAt"`
	UpdatedAt              time.Time      `json:"updatedAt"`
	ProcessedAt            *time.Time     `json:"processedAt"`
	SourcesViewed          *int64         `json:"sourcesViewed"`
	LaunchDuration         *time.Duration `json:"launchDuration"`
	ErrorMsg               *string        `json:"errorMsg"`
	DepthLevel             int            `json:"depthLevel"`
	MinWeight              float64        `json:"minWeight"`
	MaxSources             int64          `json:"maxSources"`
	MaxNeighboursForSource int64          `json:"maxNeighboursForSource"`
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

	task, err := h.tasks.GetByID(ctx, dto.ID)
	if err != nil {
		common.Error(w, err)
		return
	}

	res := dtoResponse{
		Query:                  task.Query,
		Status:                 string(task.Status),
		CreatedAt:              task.CreatedAt,
		UpdatedAt:              task.UpdatedAt,
		ProcessedAt:            task.ProcessedAt,
		DepthLevel:             task.DepthLevel,
		MinWeight:              task.MinWeight,
		MaxSources:             task.MaxSources,
		MaxNeighboursForSource: task.MaxNeighboursForSource,
	}

	if task.Status != task_model.StatusCreated && task.Status != task_model.StatusInPocessing {
		launch, err := h.launches.GetLastByTaskID(ctx, dto.ID)
		if err != nil {
			common.Error(w, err)
			return
		}

		res.SourcesViewed = &launch.SourcesViewed
		res.LaunchDuration = utils.Ptr(launch.FinishedAt.Sub(launch.StartedAt))
		res.ErrorMsg = common.ErrorSlugToMsg(launch.Error)
	}

	common.OK(w, res)
}
