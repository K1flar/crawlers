package get_protocol

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/handlers/common"
	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/internal/utils"
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
	Limit        int64   `json:"limit"`
	Offset       int64   `json:"offset"`
	TaskID       *int64  `json:"taskId"`
	Query        *string `json:"query"`
	SourceID     *int64  `json:"sourceId"`
	Title        *string `json:"title"`
	SourceStatus *string `json:"sourceStatus"`
}

type dtoResponse struct {
	Protocol []dtoProtocolItem `json:"protocol"`
}

type dtoProtocolItem struct {
	TaskID         int64          `json:"taskId"`
	Query          string         `json:"query"`
	SourceID       int64          `json:"sourceId"`
	Title          string         `json:"title"`
	URL            string         `json:"url"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	SourceStatus   string         `json:"sourceStatus"`
	LaunchID       int64          `json:"launchId"`
	LaunchNumber   int64          `json:"launchNumber"`
	StartedAt      time.Time      `json:"startedAt"`
	Duration       *time.Duration `json:"duration"`
	LaunchStatus   string         `json:"launchStatus"`
	LaunchErrorMsg *string        `json:"launchErrorMsg"`
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

	var taskQuery *string
	if dto.Query != nil {
		taskQuery = utils.Ptr(strings.ToLower(strings.Trim(*dto.Query, " ")))
	}

	var sootceTitle *string
	if dto.Title != nil {
		sootceTitle = utils.Ptr(strings.ToLower(strings.Trim(*dto.Title, " ")))
	}

	protocol, err := h.sources.GetForProtocol(ctx, storage.FilterForProtocol{
		Limit:        dto.Limit,
		Offset:       dto.Offset,
		TaskID:       dto.TaskID,
		Query:        taskQuery,
		SourceID:     dto.SourceID,
		Title:        sootceTitle,
		SourceStatus: (*source.Status)(dto.SourceStatus),
	})
	if err != nil {
		common.Error(w, err)
		return
	}

	common.OK(w, dtoResponse{
		Protocol: lo.Map(protocol, func(s source.ForProtocol, _ int) dtoProtocolItem {
			return dtoProtocolItem{
				TaskID:         s.TaskID,
				Query:          s.Query,
				SourceID:       s.SourceID,
				Title:          s.Title,
				URL:            s.URL,
				CreatedAt:      s.CreatedAt,
				UpdatedAt:      s.UpdatedAt,
				SourceStatus:   string(s.SourceStatus),
				LaunchID:       s.LaunchID,
				LaunchNumber:   s.LaunchNumber,
				StartedAt:      s.StartedAt,
				Duration:       s.Duration,
				LaunchStatus:   string(s.LaunchStatus),
				LaunchErrorMsg: common.ErrorSlugToMsg((*launch.ErrorSlug)(s.LaunchErrorSlug)),
			}
		}),
	})
}
