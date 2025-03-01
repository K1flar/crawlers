package storage

type ToCreateTask struct {
	Query string
}

type ToCreateSource struct {
	TaskID     int64
	Title      string
	URL        string
	Weight     float64
	UUID       string
	ParentUUID *string
}

type ToUpdateSource struct {
	Title *string
}
