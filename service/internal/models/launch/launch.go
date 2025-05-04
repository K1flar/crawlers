package launch

import (
	"time"
)

type Status string

type ErrorSlug string

const (
	StatusInProgress = "in_progress"
	StatusFinished   = "finished"
)

type Launch struct {
	ID            int64
	Number        int64
	TaskID        int64
	StartedAt     time.Time
	FinishedAt    *time.Time
	SourcesViewed int64
	Status        Status
	Error         *ErrorSlug
}
