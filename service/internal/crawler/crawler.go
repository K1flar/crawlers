package crawler

import (
	"context"
	"sync"

	"github.com/K1flar/crawlers/internal/gates"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
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

	existedSourcesByURL := lo.SliceToMap(sources, func(source source.Source) (string, int64) {
		return source.URL, source.ID
	})

	instance := &crawlerInstance{
		task:                task,
		searchSystem:        c.searchSystem,
		webScraper:          c.webScraper,
		sourcesStorage:      c.sourcesStorage,
		existedSourcesByURL: existedSourcesByURL,
		visited:             map[string]struct{}{},
		muVisited:           &sync.Mutex{},
		pages:               []pageInfo{},
		muPages:             &sync.Mutex{},
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
	existedSourcesByURL map[string]int64
	visited             map[string]struct{}
	muVisited           *sync.Mutex
	pages               []pageInfo
	muPages             *sync.Mutex
}

type pageInfo struct {
	IsNew      bool
	Title      *string
	URL        string
	TF         float64
	Size       int64
	UUID       string
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
	parent_uuid *string,
) error {
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
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	wg.Add(len(page.URLs))

	for _, url := range page.URLs {
		url := url
		go func() {
			defer wg.Done()
			c.crawl(ctx, url, level+1, nil)
		}()
	}

	wg.Wait()

	return nil
}
