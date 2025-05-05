package source

import "time"

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
