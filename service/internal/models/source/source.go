package source

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnavailable Status = "unavailable"
)

type Source struct {
	ID         int64
	TaskID     int64
	Title      string
	URL        string
	Status     Status
	Weight     float64
	UUID       string
	ParentUUID *string
}
