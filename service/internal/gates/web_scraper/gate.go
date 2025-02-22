package web_scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Gate struct {
	log    *slog.Logger
	client *http.Client
}

func NewGate(
	log *slog.Logger,
) *Gate {
	return &Gate{log, http.DefaultClient}
}

func (g *Gate) GetPage(ctx context.Context, url string) (*page.Page, error) {
	var err error

	page := &page.Page{}

	defer func() {
		if err != nil {
			g.log.Error(fmt.Sprintf(`error receiving the page %s: %s`, url, err))
		}
	}()

	res, err := g.client.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("source must be available")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	errGrp, _ := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		page.Title = doc.Find("title").Text()

		if page.Title == "" {
			return errors.New("title must exist")
		}

		return nil
	})

	errGrp.Go(func() error {
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			if url, exists := s.Attr("href"); exists {
				page.URLs = append(page.URLs, url)
			}
		})

		return nil
	})

	errGrp.Go(func() error {
		page.Body, err = doc.Find("body").Html()

		return err
	})

	err = errGrp.Wait()

	return page, err
}
