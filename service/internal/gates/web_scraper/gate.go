package web_scraper

import (
	"context"
	"regexp"
	"strings"
	"time"

	page_models "github.com/K1flar/crawlers/internal/models/page"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
)

const (
	defaultTimeout = 10 * time.Second

	urlPattern = `^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`
)

var (
	urlRegex = regexp.MustCompile(urlPattern)
)

type Gate struct{}

func NewGate() *Gate {
	return &Gate{}
}

func (g *Gate) GetPage(ctx context.Context, url string) (*page_models.Page, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	allocCtx, allocCtxCancel := chromedp.NewContext(ctx)
	defer allocCtxCancel()
	defer chromedp.Cancel(allocCtx)

	page := &page_models.Page{
		URL:    url,
		Status: page_models.StatusUnavailable,
	}

	var urls []string

	actions := []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.ListenTarget(ctx, func(ev any) {
				if ev, ok := ev.(*network.EventResponseReceived); ok {
					if ev.Response.URL == url || strings.HasPrefix(ev.Response.URL, url) {
						if ev.Response.Status == 200 {
							page.Status = page_models.StatusAvailable
						}
					}
				}
			})
			return nil
		}),
		chromedp.Navigate(url),
		chromedp.Title(&page.Title),
		chromedp.OuterHTML("html", &page.Body),
		// Получаем текущий URL (может отличаться от исходного из-за редиректов)
		chromedp.Location(&page.URL),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('a')).map(a => a.href);
		`, &urls),
	}

	err := chromedp.Run(allocCtx, actions...)
	if err != nil {
		return nil, err
	}

	filteredURLs := lo.Filter(urls, func(url string, _ int) bool {
		return len(url) != 0 && urlRegex.MatchString(url)
	})

	page.URLs = filteredURLs

	return page, nil
}
