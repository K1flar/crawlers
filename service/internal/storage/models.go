package storage

import (
	"time"

	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/models/source"
)

type ToCreateTask struct {
	Query                  string
	DepthLevel             int64
	MinWeight              int64
	MaxSources             int64
	MaxNeighboursForSource int64
}

type ToUpdateTask struct {
	ID                     int64
	DepthLevel             *int
	MinWeight              *float64
	MaxSources             *int64
	MaxNeighboursForSource *int64
}

type ToCreateSource struct {
	Title     string
	URL       string
	CreatedAt time.Time
	Status    source.Status
}

type ToUpdateSource struct {
	ID        int64
	Title     string
	Status    source.Status
	UpdatedAt time.Time
}

type ToCreateLaunch struct {
	TaskID    int64
	StartedAt time.Time
}

type ToFinishLaunch struct {
	ID            int64
	FinishedAt    time.Time
	SourcesViewed int64
	Error         *launch.ErrorSlug
}

type ToCreateTaskSource struct {
	TaskID         int64
	LaunchID       int64
	SourceID       int64
	ParentSourceID *int64
	Weight         float64
}
