package launcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/models/page"
	page_models "github.com/K1flar/crawlers/internal/models/page"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/services"
	"github.com/K1flar/crawlers/internal/services/collection_collector"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/internal/utils"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	log         *slog.Logger
	launches    storage.Launches
	taskSources storage.TaskSources
	sources     storage.Sources
	now         func() time.Time
}

func NewService(
	log *slog.Logger,
	launches storage.Launches,
	taskSources storage.TaskSources,
	sources storage.Sources,
) *Service {
	return &Service{
		log:         log,
		launches:    launches,
		taskSources: taskSources,
		sources:     sources,
		now:         time.Now,
	}
}

func (s *Service) Start(ctx context.Context, taskID int64) (int64, error) {
	id, err := s.launches.Create(ctx, storage.ToCreateLaunch{
		TaskID:    taskID,
		StartedAt: s.now(),
	})

	return id, err
}

func (s *Service) Finish(ctx context.Context, params services.LaunhToFinishParams) error {
	err := s.launches.Finish(ctx, storage.ToFinishLaunch{
		ID:            params.LaunchID,
		FinishedAt:    s.now(),
		SourcesViewed: int64(len(params.Pages)),
		Error:         launch.ErrorToSlug(params.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to finish launch: %w", err)
	}

	if len(params.Pages) == 0 {
		s.log.Warn(fmt.Sprintf("zero pages for task [%d], launch [%d]", params.Task.ID, params.LaunchID))

		return nil
	}

	pages := make(map[string]*page_models.PageWithParentURL, len(params.Pages))
	for url, page := range params.Pages {
		if page == nil || page.Page == nil {
			s.log.Error(fmt.Sprintf("nil page for url %s", url))
			continue
		}

		pages[url] = page
	}

	pagesWithWeight := s.calculateWeightAndFilter(pages, params.Task)

	toCreate, toUpdate, err := s.filterPages(ctx, pages, pagesWithWeight)
	if err != nil {
		return err
	}

	idByURL, err := s.createOrUpdateSources(ctx, toCreate, toUpdate)
	if err != nil {
		return err
	}

	s.log.Info(fmt.Sprintf("create %d sources, update %d sources for task %d", len(toCreate), len(toUpdate), params.Task.ID))

	err = s.taskSources.Create(ctx, s.makeParamsToCreateTaskSources(
		pagesWithWeight,
		idByURL,
		params.Task.ID, params.LaunchID,
	))

	return err
}

type pageWithWeight struct {
	URL       string
	ParentURL *string
	Weight    float64
}

func (s *Service) calculateWeightAndFilter(
	pages map[string]*page_models.PageWithParentURL,
	task task.Task,
) map[string]pageWithWeight {
	collector := collection_collector.New(task.Query)

	urls := make([]string, 0, len(pages))

	for url, page := range pages {
		if page.Status != page_models.StatusAvailable {
			continue
		}

		urls = append(urls, url)
	}

	for _, url := range urls {
		collector.AddPage(url, *pages[url].Page)
	}

	filteredPagesWithWeight := make(map[string]pageWithWeight, len(urls))
	for _, url := range urls {
		weight, ok := collector.BM25(url)
		if !ok {
			s.log.Error(fmt.Sprintf("no weight for url %s", url))
		}

		if weight < task.MinWeight {
			continue
		}

		filteredPagesWithWeight[url] = pageWithWeight{
			URL:       url,
			ParentURL: pages[url].ParentURL,
			Weight:    weight,
		}
	}

	return getConnectedPages(filteredPagesWithWeight, task.MaxSources)
}

func getConnectedPages(pages map[string]pageWithWeight, maxPages int64) map[string]pageWithWeight {
	connectivityPages := make(map[string]pageWithWeight)

	var visit func(url string)
	visit = func(url string) {
		if len(connectivityPages) >= int(maxPages) {
			return
		}

		connectivityPages[url] = pages[url]

		for u, p := range pages {
			if p.ParentURL != nil && *p.ParentURL == url {
				visit(u)
			}
		}
	}

	for url, page := range pages {
		if page.ParentURL == nil {
			visit(url)
		}
	}

	return connectivityPages
}

func (s *Service) filterPages(
	ctx context.Context,
	pages map[string]*page.PageWithParentURL,
	pagesWithWeight map[string]pageWithWeight,
) ([]storage.ToCreateSource, []storage.ToUpdateSource, error) {
	urls := lo.MapToSlice(pages, func(url string, _ *page_models.PageWithParentURL) string {
		return url
	})

	existedSourcesByURL, err := s.sources.GetByURLs(ctx, urls)
	if err != nil {
		return nil, nil, err
	}

	var (
		toUpdate []storage.ToUpdateSource
		toCreate []storage.ToCreateSource
	)

	for _, url := range urls {
		sourceStatus := source.StatusUnavailable
		if pages[url].Status == page_models.StatusAvailable {
			sourceStatus = source.StatusAvailable
		}

		if existedSource, exists := existedSourcesByURL[url]; exists {
			toUpdate = append(toUpdate, storage.ToUpdateSource{
				ID:        existedSource.ID,
				Title:     pages[url].Title,
				Status:    sourceStatus,
				UpdatedAt: s.now(),
			})
		} else {
			if pages[url].Status != page.StatusAvailable {
				continue
			}

			if _, exists := pagesWithWeight[url]; !exists {
				continue
			}

			toCreate = append(toCreate, storage.ToCreateSource{
				Title:     pages[url].Title,
				URL:       pages[url].URL,
				Status:    sourceStatus,
				CreatedAt: s.now(),
			})
		}
	}

	return toCreate, toUpdate, nil
}

func (s *Service) makeParamsToCreateTaskSources(
	pagesWithWeight map[string]pageWithWeight,
	idByURL map[string]int64,
	taskID, launchID int64,
) []storage.ToCreateTaskSource {
	res := make([]storage.ToCreateTaskSource, 0, len(pagesWithWeight))

	for url, page := range pagesWithWeight {
		var parentID *int64
		if page.ParentURL != nil {
			parentID = utils.Ptr(idByURL[*page.ParentURL])
		}

		res = append(res, storage.ToCreateTaskSource{
			TaskID:         taskID,
			LaunchID:       launchID,
			SourceID:       idByURL[url],
			ParentSourceID: parentID,
			Weight:         page.Weight,
		})
	}

	return res
}

func (s *Service) createOrUpdateSources(
	ctx context.Context,
	toCreate []storage.ToCreateSource,
	toUpdate []storage.ToUpdateSource,
) (map[string]int64, error) {
	var (
		created map[string]int64
		updated map[string]int64
		err     error
	)

	errGrp, gCtx := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		created, err = s.sources.Create(gCtx, toCreate)
		if err != nil {
			return err
		}

		if len(created) != len(toCreate) {
			return fmt.Errorf("created %d sources (must be %d)", len(created), len(toCreate))
		}

		return nil
	})

	errGrp.Go(func() error {
		updated, err = s.sources.Update(gCtx, toUpdate)
		if err != nil {
			return err
		}

		if len(updated) != len(toUpdate) {
			return fmt.Errorf("updated %d sources (must be %d)", len(updated), len(toUpdate))
		}

		return nil
	})

	err = errGrp.Wait()
	if err != nil {
		return nil, err
	}

	res := make(map[string]int64, len(toCreate)+len(toUpdate))

	for url, id := range created {
		res[url] = id
	}

	for url, id := range updated {
		res[url] = id
	}

	if len(res) != len(toCreate)+len(toUpdate) {
		return nil, errors.New("intersection created and updated sources")
	}

	return res, nil
}
