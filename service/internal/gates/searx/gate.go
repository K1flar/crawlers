package searx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

type Gate struct {
	log    *slog.Logger
	client Client
}

func NewGate(
	log *slog.Logger,
	client Client,
) *Gate {
	return &Gate{log, client}
}

type dtoSearchResponse struct {
	Res []dtoSearchSource `json:"results"`
}

type dtoSearchSource struct {
	URL string `json:"url"`
}

func (g *Gate) Search(ctx context.Context, query string) ([]string, error) {
	var err error

	defer func() {
		if err != nil {
			g.log.Error(fmt.Sprintf(`error to search "%s": %s`, query, err))
		}
	}()

	values := url.Values{}
	values.Set("q", query)
	values.Set("format", "json")

	url := &url.URL{
		Path:     "/search",
		RawQuery: values.Encode(),
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
	}

	res, err := g.client.Do(ctx, req)
	if err != nil {
		return []string{}, err
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return []string{}, err
	}
	defer res.Body.Close()

	var dtoResponse dtoSearchResponse
	err = json.Unmarshal(b, &dtoResponse)
	if err != nil {
		return []string{}, err
	}

	urls := make([]string, 0, len(dtoResponse.Res))
	for _, source := range dtoResponse.Res {
		if strings.Contains(source.URL, "http") {
			urls = append(urls, source.URL)
		}
	}

	return urls, nil
}
