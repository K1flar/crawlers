package web_scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"

type Gate struct {
	log    *slog.Logger
	client *http.Client
}

func NewGate(
	log *slog.Logger,
) *Gate {
	return &Gate{log, http.DefaultClient}
}

func (g *Gate) GetPage(ctx context.Context, url string) (page.Page, error) {
	var err error

	page := page.Page{}

	defer func() {
		if err != nil {
			g.log.Error(fmt.Sprintf(`error receiving the page %s: %s`, url, err))
		}
	}()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return page, err
	}

	req.Header.Set("User-Agent", defaultUserAgent)

	res, err := g.client.Do(req)
	if err != nil {
		return page, err
	}

	if res.StatusCode != http.StatusOK {
		return page, business_errors.UnavailableSource
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return page, err
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
