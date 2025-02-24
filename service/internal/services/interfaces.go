package services

import "github.com/K1flar/crawlers/internal/models/page"

type CollectionCollector interface {
	AddPage(uuid string, page page.Page)
	BM25(uuid string) (float64, bool)
}
