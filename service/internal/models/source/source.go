package source

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnacailable Status = "unavailable"
)

type Source struct {
	ID     int64
	Title  string
	URL    string
	Status Status
	Weight float64
}
