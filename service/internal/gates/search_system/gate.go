package search_system

import (
	"context"
	"log/slog"
)

type Gate struct {
	log *slog.Logger
}

func NewGate(
	log *slog.Logger,
) *Gate {
	return &Gate{log}
}

func (g *Gate) Search(ctx context.Context, query string) ([]string, error) {
	urls := []string{}

	return urls, nil
}
