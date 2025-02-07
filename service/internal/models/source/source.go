package source

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnacailable Status = "unavailable"
)

type Source struct {
	ID     int64
	URL    string
	Status Status
	Weight float64
}
