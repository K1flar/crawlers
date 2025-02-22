package crawler

import (
	"context"
	"sync"

	"github.com/K1flar/crawlers/internal/gates"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
)

type Crawler struct {
	task           task.Task
	searchSystem   gates.SearchSystem
	sourcesStorage storage.Sources
	visited        map[string]struct{}
	mu             *sync.RWMutex
}

func New(
	task task.Task,
	searchSystem gates.SearchSystem,
	sourcesStorage storage.Sources,
) *Crawler {
	return &Crawler{
		task:           task,
		searchSystem:   searchSystem,
		sourcesStorage: sourcesStorage,
		visited:        map[string]struct{}{},
		mu:             &sync.RWMutex{},
	}
}

func (c *Crawler) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}

	urls, err := c.searchSystem.Search(ctx, c.task.Query)
	if err != nil {
		return err
	}

	wg.Add(len(urls))

	for _, url := range urls {
		url := url
		go func() {
			defer wg.Done()
			c.crawl(ctx, url, 0)
		}()
	}

	wg.Wait()

	return nil
}

func (c *Crawler) crawl(ctx context.Context, url string, level int) error {

	return nil
}
