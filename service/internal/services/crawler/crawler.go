package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/gates"
	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/gammazero/workerpool"
)

const (
	countWorkers = 5
)

type Crawler struct {
	log          *slog.Logger
	searchSystem gates.SearchSystem
	webScraper   gates.WebScraper
	now          func() time.Time
}

func New(
	log *slog.Logger,
	searchSystem gates.SearchSystem,
	webScraper gates.WebScraper,
) *Crawler {
	return &Crawler{
		log:          log,
		searchSystem: searchSystem,
		webScraper:   webScraper,
		now:          time.Now,
	}
}

func (c *Crawler) Start(ctx context.Context, task task.Task) (map[string]*page.PageWithParentURL, error) {
	instance := c.newInstance(task)

	urls, err := c.searchSystem.Search(ctx, task.Query)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to search initial sources by query [%s]: %w", business_errors.SearxError, task.Query, err)
	}

	if len(urls) == 0 {
		return nil, business_errors.ZeroStartSources
	}

	timeStart := c.now()
	c.log.Info(fmt.Sprintf("start crawler for task [%d]: [%s]", task.ID, task.Query))

	err = instance.start(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("failed to run crawler instance: %w", err)
	}

	c.log.Info(fmt.Sprintf("end search for task [%d] with %d sources. %s", task.ID, len(instance.pages), time.Since(timeStart)))

	return instance.pages, nil
}

func (c *Crawler) newInstance(task task.Task) *crawlerInstance {
	return &crawlerInstance{
		task:         task,
		webScraper:   c.webScraper,
		wp:           workerpool.New(countWorkers),
		stop:         make(chan struct{}),
		pending:      0,
		pendingLock:  &sync.Mutex{},
		crawlerTasks: make(chan crawlerTask),
		visited:      make(map[string]struct{}, task.MaxSources),
		pages:        make(map[string]*page.PageWithParentURL, task.MaxSources),
	}
}
