package crawler

import (
	"context"
	"log/slog"

	"github.com/K1flar/crawlers/internal/storage"
)

type Crawler struct {
	log            *slog.Logger
	sourcesStorage storage.Sources
}

func New(
	log *slog.Logger,
	sourcesStorage storage.Sources,
) *Crawler {
	return &Crawler{log, sourcesStorage}
}

func (c *Crawler) Start(ctx context.Context, query string) error {

	return nil
}
