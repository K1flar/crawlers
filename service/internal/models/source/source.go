package source

import (
	"time"

	"github.com/K1flar/crawlers/internal/models/launch"
)

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnavailable Status = "unavailable"
)

type Source struct {
	ID        int64
	URL       string
	Title     string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ForTask struct {
	ID       int64
	URL      string
	Title    string
	Weight   float64
	ParentID *int64
}

type ForProtocol struct {
	TaskID          int64
	Query           string
	SourceID        int64
	Title           string
	URL             string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	SourceStatus    Status
	LaunchID        int64
	LaunchNumber    int64
	StartedAt       time.Time
	Duration        *time.Duration
	LaunchStatus    launch.Status
	LaunchErrorSlug *string
}
