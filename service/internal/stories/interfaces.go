package stories

import "context"

type CreateTask interface {
	Create(ctx context.Context, query string) (int64, error)
}
