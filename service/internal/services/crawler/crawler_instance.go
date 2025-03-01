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
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/utils"
	"github.com/google/uuid"
)

type crawlerInstance struct {
	task                task.Task
	searchSystem        gates.SearchSystem
	webScraper          gates.WebScraper
	sourcesStorage      storage.Sources
	collectionCollector services.CollectionCollector
	existedSourcesByURL map[string]string
	visited             map[string]struct{}
	visitedLock         *sync.Mutex
	sourcesToCreate     []sourceToCreate
	sourcesToCreateLock *sync.Mutex
	sourcesToUpdate     []sourceToUpdate
	sourcesToUpdateLock *sync.Mutex
}

type sourceToCreate struct {
	UUID       string
	Title      string
	URL        string
	ParentUUID *string
}

type sourceToUpdate struct {
	UUID   string
	Title  *string
	Status *source.Status
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
	if level > c.task.DepthLevel {
		return nil
	}

	c.visitedLock.Lock()
	if _, exists := c.visited[url]; exists {
		c.visitedLock.Unlock()
		return nil
	}
	c.visited[url] = struct{}{}
	c.visitedLock.Unlock()

	sourceUUID, sourceExists := c.existedSourcesByURL[url]
	if !sourceExists {
		sourceUUID = uuid.NewString()
	}

	page, err := c.webScraper.GetPage(ctx, url)
	switch {
	case errors.Is(err, business_errors.UnavailableSource) && sourceExists:
		c.addPageToUpdate(sourceToUpdate{UUID: sourceUUID, Status: utils.Ptr(source.StatusUnavailable)})
		return nil
	case err != nil:
		return err
	}

	if sourceExists {
		c.addPageToUpdate(sourceToUpdate{
			UUID:   sourceUUID,
			Title:  &page.Title,
			Status: utils.Ptr(source.StatusAvailable),
		})
	} else {
		c.addPageToCreate(sourceToCreate{
			UUID:       sourceUUID,
			Title:      page.Title,
			URL:        url,
			ParentUUID: parentUUID,
		})
	}

	c.collectionCollector.AddPage(sourceUUID, page)

	wg := &sync.WaitGroup{}

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

func (c *crawlerInstance) addPageToCreate(page sourceToCreate) {
	c.sourcesToCreateLock.Lock()
	c.sourcesToCreate = append(c.sourcesToCreate, page)
	c.sourcesToCreateLock.Unlock()
}

func (c *crawlerInstance) addPageToUpdate(page sourceToUpdate) {
	c.sourcesToUpdateLock.Lock()
	c.sourcesToUpdate = append(c.sourcesToUpdate, page)
	c.sourcesToUpdateLock.Unlock()
}
