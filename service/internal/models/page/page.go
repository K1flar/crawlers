package page

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnavailable Status = "unavailable"
)

type Page struct {
	URL     string
	Status  Status
	Title   string
	Content string
	URLs    []string
}

type PageWithParentURL struct {
	ParentURL *string
	*Page
}
