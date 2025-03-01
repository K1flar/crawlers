package http_client

import (
	"context"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	opts       []Option
}

func New(opts ...Option) *Client {
	return &Client{http.DefaultClient, opts}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req = req.WithContext(ctx)

	for _, opt := range c.opts {
		opt.PrepareRequest(req)
	}

	return c.httpClient.Do(req)
}

type Option interface {
	PrepareRequest(*http.Request)
}

type withBaseURLOpt struct {
	baseURL string
}

func (o *withBaseURLOpt) PrepareRequest(req *http.Request) {
	req.URL.Host = o.baseURL
}

func WithBaseURL(baseURL string) Option {
	return &withBaseURLOpt{baseURL}
}
