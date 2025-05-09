package get_tasks

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/K1flar/crawlers/internal/handlers/common"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/internal/utils"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
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
	ID     int64   `json:"id"`
	Limit  int64   `json:"limit"`
	Offset int64   `json:"offset"`
	Status *string `json:"status"`
	Query  *string `json:"query"`
}

type dtoResponse struct {
	Tasks []dtoTask `json:"tasks"`
	Total int64     `json:"total"`
}

type dtoTask struct {
	ID           int64  `json:"id"`
	Query        string `json:"query"`
	Status       string `json:"status"`
	CountSources int64  `json:"countSources"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err   error
		tasks []task.ForList
		count int64
	)

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

	var query *string
	if dto.Query != nil {
		query = utils.Ptr(strings.ToLower(strings.Trim(*dto.Query, " ")))
	}

	errGrp, gCtx := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		tasks, err = h.tasks.GetForList(gCtx, storage.FilterTaskForList{
			Limit:  dto.Limit,
			Offset: dto.Offset,
			Status: (*task.Status)(dto.Status),
			Query:  query,
		})

		return err
	})

	errGrp.Go(func() error {
		count, err = h.tasks.GetCount(gCtx)

		return err
	})

	err = errGrp.Wait()
	if err != nil {
		common.Error(w, err)
		return
	}

	common.OK(w, dtoResponse{
		Tasks: lo.Map(tasks, func(task task.ForList, _ int) dtoTask {
			return dtoTask{
				ID:           task.ID,
				Query:        task.Query,
				Status:       string(task.Status),
				CountSources: task.CountSources,
			}
		}),
		Total: count,
	})
}
