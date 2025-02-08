package task

import "time"

type Status string

const (
	StatusCreated  Status = "created"
	StatusActive   Status = "active"
	StatusStopped  Status = "stopped"
	StatusInactive Status = "inactive"
)

type Task struct {
	ID         int64
	Query      string
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DepthLevel int
	MinWeight  float64
}
