package document

type Document struct {
	UUID string
	Size int64
	TF   map[string]float64
}
