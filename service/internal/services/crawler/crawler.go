package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/K1flar/crawlers/internal/gates"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/services/collection_collector"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/samber/lo"
)

type Crawler struct {
	log            *slog.Logger
	searchSystem   gates.SearchSystem
	webScraper     gates.WebScraper
	sourcesStorage storage.Sources
	now            func() time.Time
}

func New(
	log *slog.Logger,
	searchSystem gates.SearchSystem,
	webScraper gates.WebScraper,
	sourcesStorage storage.Sources,
) *Crawler {
	return &Crawler{
		log:            log,
		searchSystem:   searchSystem,
		webScraper:     webScraper,
		sourcesStorage: sourcesStorage,
		now:            time.Now,
	}
}

// Алгоритм:
// 1. Получение существующих источников по задаче (для определения обновления/создания источников из обхода)
// 2. Запуск рекурсивного обхода источников, сохранение параметров каждой страницы
// 3. Подсчет релевантности каждой страницы по алгоритму BM25
// 4. Обновление/создание источников
// 5. Удаление незадействованных источников
func (c *Crawler) Start(ctx context.Context, task task.Task) error {
	sources, err := c.sourcesStorage.GetByTaskID(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to get task by ID %d: %w", task.ID, err)
	}

	instance := c.newInstance(task, sources)

	timeStart := c.now()
	c.log.Info(fmt.Sprintf("start crawler for task [%d]: [%s]", task.ID, task.Query))

	err = instance.start(ctx)
	if err != nil {
		return fmt.Errorf("failed to run crawler instance: %w", err)
	}

	c.log.Info(fmt.Sprintf("end search for task [%d] with %d sources to create and %d sources to update, [%s]",
		task.ID, len(instance.sourcesToCreate), len(instance.sourcesToUpdate), time.Since(timeStart)))

	_, err = c.createSources(ctx, instance)
	if err != nil {
		return fmt.Errorf("failed to create sources: %w", err)
	}

	return err
}

func (c *Crawler) newInstance(task task.Task, existedSources []source.Source) *crawlerInstance {
	existedSourcesByURL := lo.SliceToMap(existedSources, func(source source.Source) (string, string) {
		return source.URL, source.UUID
	})

	return &crawlerInstance{
		task:                task,
		searchSystem:        c.searchSystem,
		webScraper:          c.webScraper,
		sourcesStorage:      c.sourcesStorage,
		collectionCollector: collection_collector.New(task.Query),
		existedSourcesByURL: existedSourcesByURL,
		visited:             map[string]struct{}{},
		visitedLock:         &sync.Mutex{},
		sourcesToCreate:     []sourceToCreate{},
		sourcesToCreateLock: &sync.Mutex{},
		sourcesToUpdate:     []sourceToUpdate{},
		sourcesToUpdateLock: &sync.Mutex{},
	}
}

func (c *Crawler) createSources(ctx context.Context, instance *crawlerInstance) ([]int64, error) {
	sourcesToCreate := make([]storage.ToCreateSource, 0, len(instance.sourcesToCreate))

	for _, page := range instance.sourcesToCreate {
		bm25, ok := instance.collectionCollector.BM25(page.UUID)
		if !ok || bm25 <= instance.task.MinWeight {
			continue
		}

		sourcesToCreate = append(sourcesToCreate, storage.ToCreateSource{
			TaskID:     instance.task.ID,
			Title:      page.Title,
			URL:        page.URL,
			Weight:     bm25,
			UUID:       page.UUID,
			ParentUUID: page.ParentUUID,
		})
	}

	return c.sourcesStorage.Create(ctx, sourcesToCreate)
}
