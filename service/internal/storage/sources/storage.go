package sources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

var _ storage.Sources = (*Storage)(nil)

type Storage struct {
	db *sqlx.DB
}

var pgSql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	sourcesTbl = "sources"

	idCol        = "id"
	titleCol     = "title"
	urlCol       = "url"
	statusCol    = "status"
	createdAtCol = "created_at"
	updatedAtCol = "updated_at"
)

var readColumns = []string{idCol, titleCol, urlCol, statusCol, createdAtCol, updatedAtCol}

type sourcePG struct {
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	URL       string    `db:"url"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
}

type key struct {
	URL string `db:"url"`
	ID  int64  `db:"id"`
}

func (s *Storage) Create(ctx context.Context, params []storage.ToCreateSource) (map[string]int64, error) {
	if len(params) == 0 {
		return map[string]int64{}, nil
	}

	q := pgSql.
		Insert(sourcesTbl).
		Columns(titleCol, urlCol, statusCol, createdAtCol, updatedAtCol)

	for _, p := range params {
		q = q.Values(p.Title, p.URL, p.Status, p.CreatedAt, p.CreatedAt)
	}

	sql, args := q.Suffix(returning(urlCol, idCol)).MustSql()

	var out []key

	err := s.db.SelectContext(ctx, &out, sql, args...)

	return lo.SliceToMap(out, func(key key) (string, int64) {
		return key.URL, key.ID
	}), err
}

func (s *Storage) Update(ctx context.Context, params []storage.ToUpdateSource) (map[string]int64, error) {
	if len(params) == 0 {
		return map[string]int64{}, nil
	}

	out := make(map[string]int64, len(params))

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, param := range params {
		sql, args := pgSql.
			Update(sourcesTbl).
			Set(titleCol, param.Title).
			Set(statusCol, param.Status).
			Set(updatedAtCol, param.UpdatedAt).
			Where(squirrel.Eq{idCol: param.ID}).
			Suffix(returning(urlCol, idCol)).
			MustSql()

		var key key
		err := tx.GetContext(ctx, &key, sql, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute update query: %w", err)
		}

		out[key.URL] = key.ID
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return out, nil
}

func (s *Storage) GetByURLs(ctx context.Context, urls []string) (map[string]source.Source, error) {
	if len(urls) == 0 {
		return map[string]source.Source{}, nil
	}

	var res []sourcePG

	sql, args := pgSql.
		Select(readColumns...).
		From(sourcesTbl).
		Where(squirrel.Eq{urlCol: urls}).
		MustSql()

	err := s.db.SelectContext(ctx, &res, sql, args...)
	if err != nil {
		return nil, err
	}

	return lo.SliceToMap(res, func(source sourcePG) (string, source.Source) {
		return source.URL, mapFromPG(source)
	}), nil
}

func mapFromPGMany(sources []sourcePG) []source.Source {
	return lo.Map(sources, func(pg sourcePG, _ int) source.Source {
		return mapFromPG(pg)
	})
}

func mapFromPG(pg sourcePG) source.Source {
	return source.Source{
		ID:        pg.ID,
		Title:     pg.Title,
		URL:       pg.URL,
		Status:    source.Status(pg.Status),
		CreatedAt: pg.CreatedAt,
		UpdatedAt: pg.UpdatedAt,
	}
}

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
