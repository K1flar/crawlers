package web_scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/models/page"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"
const defaultTimeout = 2 * time.Second

type Gate struct {
	client *http.Client
}

func NewGate() *Gate {
	return &Gate{http.DefaultClient}
}

func (g *Gate) GetPage(ctx context.Context, url string) (page.Page, error) {
	page := page.Page{}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return page, err
	}

	req.Header.Set("User-Agent", defaultUserAgent)

	res, err := g.client.Do(req)
	if err != nil {
		return page, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return page, business_errors.UnavailableSource
	}

	var reader io.Reader

	_, params, _ := mime.ParseMediaType(res.Header.Get("Content-type"))
	contentType := strings.ToLower(params["charset"])

	switch contentType {
	case "utf-8":
		reader = res.Body
	case "windows-1251":
		reader = charmap.Windows1251.NewDecoder().Reader(res.Body)
	case "iso-8859-1":
		reader = charmap.ISO8859_1.NewDecoder().Reader(res.Body)
	default:
		return page, fmt.Errorf("unsopported content type")
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return page, err
	}

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
