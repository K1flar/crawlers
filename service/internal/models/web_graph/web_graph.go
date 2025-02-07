package web_graph

import "github.com/K1flar/crawlers/internal/models/source"

type WebGraph struct {
	Pages []WebPage
}

type WebPage struct {
	source.Source
	Pages []WebPage
}
