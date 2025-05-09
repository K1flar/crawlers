package document

type Document struct {
	URL  string
	Size int64
	TF   map[string]int64
}
