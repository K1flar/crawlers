package gates

import "context"

type SearchSystem interface {
	Search(ctx context.Context, query string) ([]string, error)
}
