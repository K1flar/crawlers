package task

import "time"

type Status string

const (
	StatusCreated          Status = "created"
	StatusActive           Status = "active"
	StatusInPocessing      Status = "in_processing"
	StatusStopped          Status = "stopped"
	StatusStoppedWithError Status = "stopped_with_error"
	StatusInactive         Status = "inactive"
)

type Task struct {
	ID                     int64
	Query                  string
	Status                 Status
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ProcessedAt            *time.Time
	DepthLevel             int
	MinWeight              float64
	MaxSources             int64
	MaxNeighboursForSource int64
}

type ForList struct {
	ID           int64
	Query        string
	Status       Status
	CountSources int64
}
