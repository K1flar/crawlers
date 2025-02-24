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
	collection map[string]document.Document // коллекция документов по UUID
	df         map[string]int               // количество документов, содержащих определенный термин
	avgSize    float64                      // средний размер документов
	totalSize  int64                        // общий размер слов в коллекции
}

func New(q string) *Service {
	terms := strings.Split(q, " ")

	return &Service{
		terms: lo.SliceToMap(terms, func(term string) (string, struct{}) {
			return term, struct{}{}
		}),
		collection: map[string]document.Document{},
		df:         map[string]int{},
	}
}

func (s *Service) AddPage(uuid string, page page.Page) {
	words := strings.Split(page.Body, " ")
	size := int64(len(words))

	termsCount := make(map[string]int64, len(s.terms))
	for _, word := range words {
		if _, ok := s.terms[word]; ok {
			termsCount[word]++
		}
	}

	tf := make(map[string]float64, len(s.terms))
	for term, count := range termsCount {
		tf[term] = float64(count / size)
	}

	s.collection[uuid] = document.Document{
		UUID: uuid,
		Size: size,
		TF:   tf,
	}

	s.totalSize += size
	s.avgSize = float64(s.totalSize / int64(len(s.collection)))
}

// idf https://habr.com/ru/articles/840268/
func (s *Service) BM25(uuid string) (float64, bool) {
	doc, ok := s.collection[uuid]
	if !ok {
		return 0, false
	}

	bm25 := float64(0)

	for term := range s.terms {
		numerator := float64(doc.TF[term] * (k + 1))
		denominator := float64(doc.TF[term]+k*(1-b+b*float64(doc.Size)/s.avgSize)) + 0.5
		bm25 += s.idf(term) * numerator / denominator
	}

	return bm25, true
}

func (s *Service) idf(term string) float64 {
	numerator := float64(len(s.collection)-s.df[term]) + 0.5
	denominator := float64(s.df[term]) + 0.5
	return math.Log(numerator/denominator + 1)
}
