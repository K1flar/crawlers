package gates

import (
	"context"

	"github.com/K1flar/crawlers/internal/models/page"
)

type SearchSystem interface {
	Search(ctx context.Context, query string) ([]string, error)
}

type WebScraper interface {
	GetPage(ctx context.Context, url string) (page.Page, error)
}
