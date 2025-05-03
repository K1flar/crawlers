package crawler

import (
	"context"
	"fmt"
	"sync"

	"github.com/K1flar/crawlers/internal/gates"
	page_models "github.com/K1flar/crawlers/internal/models/page"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/gammazero/workerpool"
	"github.com/samber/lo"
)

type crawlerTask struct {
	DepthLevel int
	Page       *page_models.PageWithParentURL
}

type crawlerInstance struct {
	task         task.Task
	webScraper   gates.WebScraper
	wp           *workerpool.WorkerPool
	crawlerTasks chan crawlerTask
	stop         chan struct{}
	pending      int64
	pendingLock  *sync.Mutex
	visited      map[string]struct{}
	pages        map[string]*page_models.PageWithParentURL
}

func (c *crawlerInstance) start(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	fmt.Println("start urls: ", len(urls))

	go c.crawl(ctx)

	for _, url := range urls {
		url := url
		c.incPending()
		c.wp.Submit(func() {
			fmt.Println("get ", url)
			page, _ := c.webScraper.GetPage(ctx, url)

			c.crawlerTasks <- crawlerTask{
				DepthLevel: 1,
				Page:       &page_models.PageWithParentURL{Page: page},
			}
		})
	}

	<-c.stop
	c.wp.Stop()
	close(c.crawlerTasks)

	return nil
}

func (c *crawlerInstance) crawl(ctx context.Context) {
	for task := range c.crawlerTasks {
		c.decPending()

		if task.Page == nil || task.Page.Page == nil {
			continue
		}

		if _, visited := c.visited[task.Page.URL]; visited {
			continue
		}

		c.visited[task.Page.URL] = struct{}{}
		c.pages[task.Page.URL] = task.Page

		if task.DepthLevel >= c.task.DepthLevel {
			continue
		}

		if len(c.pages) > int(c.task.MaxSources) {
			select {
			case c.stop <- struct{}{}:
			default:
			}
			continue
		}

		urls := c.filterURLs(task.Page.URLs)

		if len(urls) == 0 {
			continue
		}

		for _, url := range urls {
			url := url
			c.incPending()
			c.wp.Submit(func() {
				page, _ := c.webScraper.GetPage(ctx, url)

				c.crawlerTasks <- crawlerTask{
					DepthLevel: task.DepthLevel + 1,
					Page: &page_models.PageWithParentURL{
						ParentURL: &task.Page.URL,
						Page:      page,
					},
				}
			})
		}

	}
}

func (c *crawlerInstance) filterURLs(urls []string) []string {
	notVisited := lo.Filter(urls, func(url string, _ int) bool {
		_, visited := c.visited[url]
		return !visited
	})

	if len(notVisited) > int(c.task.MaxNeighboursForSource) {
		notVisited = notVisited[:c.task.MaxNeighboursForSource]
	}

	return notVisited
}

func (c *crawlerInstance) incPending() {
	c.pendingLock.Lock()
	c.pending++
	c.pendingLock.Unlock()
}

func (c *crawlerInstance) decPending() {
	c.pendingLock.Lock()
	c.pending--
	if c.pending == 0 {
		select {
		case c.stop <- struct{}{}:
		default:
		}
	}
	c.pendingLock.Unlock()
}
