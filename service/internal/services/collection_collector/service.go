package collection_collector

import (
	"math"
	"strings"

	"github.com/K1flar/crawlers/internal/models/document"
	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/samber/lo"
)

const (
	// Гиперпараметры для алгоритма BM25
	k = 1.2
	b = 0.75
)

type Service struct {
	terms      map[string]struct{}          // список терминов
	collection map[string]document.Document // коллекция документов по URL
	df         map[string]int               // количество документов, содержащих определенный термин
	avgSize    float64                      // средний размер документов
	totalSize  int64                        // общий размер слов в коллекции
}

func New(q string) *Service {
	terms := lo.SliceToMap(strings.Split(strings.ToLower(q), " "), func(term string) (string, struct{}) {
		return term, struct{}{}
	})

	return &Service{
		terms:      terms,
		collection: map[string]document.Document{},
		df:         make(map[string]int, len(terms)),
	}
}

func (s *Service) AddPage(url string, page page.Page) {
	words := strings.Fields(page.Content)
	docLength := int64(len(words))

	tf := make(map[string]int64, len(s.terms))
	for _, word := range words {
		word = strings.ToLower(word)
		if _, ok := s.terms[word]; ok {
			tf[word]++
		}
	}

	s.collection[url] = document.Document{
		URL:  url,
		Size: docLength,
		TF:   tf,
	}

	s.totalSize += docLength
	s.avgSize = float64(s.totalSize) / float64(len(s.collection))

	for term, count := range tf {
		if count > 0 {
			s.df[term]++
		}
	}
}

// idf https://habr.com/ru/articles/840268/
func (s *Service) BM25(url string) (float64, bool) {
	doc, ok := s.collection[url]
	if !ok {
		return 0, false
	}

	totalDocs := len(s.collection)
	bm25 := float64(0)

	for term := range s.terms {
		tf := float64(doc.TF[term])
		df := float64(s.df[term])

		idf := math.Log(float64(totalDocs)-df+0.5)/(df+0.5) + 1.0

		numerator := tf * (k + 1)
		denominator := tf + k*(1-b+b*float64(doc.Size)/s.avgSize)
		bm25 += idf * numerator / denominator
	}

	return bm25, true
}
