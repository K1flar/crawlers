package crawler

import (
	"context"
	"errors"
	"sync"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/gates"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/services"
	"github.com/K1flar/crawlers/internal/services/collection_collector"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/utils"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type Crawler struct {
	searchSystem   gates.SearchSystem
	webScraper     gates.WebScraper
	sourcesStorage storage.Sources
}

func New(
	searchSystem gates.SearchSystem,
	webScraper gates.WebScraper,
	sourcesStorage storage.Sources,
) *Crawler {
	return &Crawler{
		searchSystem:   searchSystem,
		webScraper:     webScraper,
		sourcesStorage: sourcesStorage,
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
		return err
	}

	existedSourcesByURL := lo.SliceToMap(sources, func(source source.Source) (string, string) {
		return source.URL, source.UUID
	})

	instance := &crawlerInstance{
		task:                task,
		searchSystem:        c.searchSystem,
		webScraper:          c.webScraper,
		sourcesStorage:      c.sourcesStorage,
		collectionCollector: collection_collector.New(task.Query),
		existedSourcesByURL: existedSourcesByURL,
		visited:             map[string]struct{}{},
		muVisited:           &sync.Mutex{},
		pagesToUpdate:       []pageToUpdate{},
		muPagesToUpdate:     &sync.Mutex{},
		pagesToCreate:       []pageToCreate{},
		muPagesToCreate:     &sync.Mutex{},
	}

	err = instance.start(ctx)
	if err != nil {
		return err
	}

	return nil
}

type crawlerInstance struct {
	task                task.Task
	searchSystem        gates.SearchSystem
	webScraper          gates.WebScraper
	sourcesStorage      storage.Sources
	collectionCollector services.CollectionCollector
	existedSourcesByURL map[string]string
	visited             map[string]struct{}
	muVisited           *sync.Mutex
	pagesToUpdate       []pageToUpdate
	muPagesToUpdate     *sync.Mutex
	pagesToCreate       []pageToCreate
	muPagesToCreate     *sync.Mutex
}

type pageToUpdate struct {
	UUID   string
	Title  *string
	Status *source.Status
}

type pageToCreate struct {
	UUID       string
	Title      string
	URL        string
	ParentUUID *string
}

func (c *crawlerInstance) start(ctx context.Context) error {
	urls, err := c.searchSystem.Search(ctx, c.task.Query)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(urls))

	for _, url := range urls {
		url := url
		go func() {
			defer wg.Done()
			c.crawl(ctx, url, 0, nil)
		}()
	}

	wg.Wait()

	return nil
}

func (c *crawlerInstance) crawl(
	ctx context.Context,
	url string,
	level int,
	parentUUID *string,
) error {
	sourceUUID, sourceExists := c.existedSourcesByURL[url]
	if !sourceExists {
		sourceUUID = uuid.NewString()
	}

	if level > c.task.DepthLevel {
		return nil
	}

	c.muVisited.Lock()
	if _, exists := c.visited[url]; exists {
		c.muVisited.Unlock()
		return nil
	}
	c.visited[url] = struct{}{}
	c.muVisited.Unlock()

	page, err := c.webScraper.GetPage(ctx, url)
	switch {
	case errors.Is(err, business_errors.UnavailableSource) && sourceExists:
		c.addPageToUpdate(pageToUpdate{UUID: sourceUUID, Status: utils.Ptr(source.StatusUnavailable)})
		return nil
	case err != nil:
		return err
	}

	if sourceExists {
		c.addPageToUpdate(pageToUpdate{
			UUID:   sourceUUID,
			Title:  &page.Title,
			Status: utils.Ptr(source.StatusUnavailable),
		})
	} else {
		c.addPageToCreate(pageToCreate{
			UUID:       sourceUUID,
			Title:      page.Title,
			URL:        url,
			ParentUUID: parentUUID,
		})
	}

	c.collectionCollector.AddPage(sourceUUID, page)

	wg := sync.WaitGroup{}

	wg.Add(len(page.URLs))

	for _, url := range page.URLs {
		url := url
		go func() {
			defer wg.Done()
			c.crawl(ctx, url, level+1, &sourceUUID)
		}()
	}

	wg.Wait()

	return nil
}

func (c *crawlerInstance) addPageToUpdate(page pageToUpdate) {
	c.muPagesToUpdate.Lock()
	c.pagesToUpdate = append(c.pagesToUpdate, page)
	c.muPagesToUpdate.Unlock()
}

func (c *crawlerInstance) addPageToCreate(page pageToCreate) {
	c.muPagesToCreate.Lock()
	c.pagesToCreate = append(c.pagesToCreate, page)
	c.muPagesToCreate.Unlock()
}
