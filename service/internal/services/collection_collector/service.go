package collection_collector

import (
	"math"
	"strings"
	"sync"

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
	mu         *sync.RWMutex
}

func New(q string) *Service {
	terms := lo.SliceToMap(strings.Split(strings.ToLower(q), " "), func(term string) (string, struct{}) {
		return term, struct{}{}
	})

	return &Service{
		terms:      terms,
		collection: map[string]document.Document{},
		df:         make(map[string]int, len(terms)),
		mu:         &sync.RWMutex{},
	}
}

func (s *Service) AddPage(url string, page page.Page) {
	words := strings.Split(page.Body, " ")
	size := int64(len(words))

	termsCount := make(map[string]int64, len(s.terms))
	for _, word := range words {
		word = strings.ToLower(word)
		if _, ok := s.terms[word]; ok {
			termsCount[word]++
		}
	}

	tf := make(map[string]float64, len(s.terms))
	for term := range s.terms {
		tf[term] = float64(termsCount[term]) / float64(size)
	}

	s.mu.Lock()
	s.collection[url] = document.Document{
		URL:  url,
		Size: size,
		TF:   tf,
	}

	s.totalSize += size
	s.avgSize = float64(s.totalSize) / float64(len(s.collection))

	for term := range s.terms {
		if termsCount[term] > 0 {
			s.df[term]++
		}
	}
	s.mu.Unlock()
}

// idf https://habr.com/ru/articles/840268/
func (s *Service) BM25(url string) (float64, bool) {
	doc, ok := s.collection[url]
	if !ok {
		return 0, false
	}

	bm25 := float64(0)

	for term := range s.terms {
		numerator := float64(doc.TF[term] * (k + 1))
		denominator := float64(doc.TF[term] + k*(1-b+b*float64(doc.Size)/s.avgSize))
		bm25 += s.idf(term) * numerator / denominator
	}

	return bm25, true
}

func (s *Service) idf(term string) float64 {
	numerator := float64(len(s.collection)-s.df[term]) + 0.5
	denominator := float64(s.df[term]) + 0.5
	return math.Log(numerator/denominator + 1)
}
