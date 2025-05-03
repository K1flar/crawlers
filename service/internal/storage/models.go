package storage

type ToCreateTask struct {
	Query                  string
	DepthLevel             int64
	MinWeight              int64
	MaxSources             int64
	MaxNeighboursForSource int64
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
