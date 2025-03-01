package sources

import (
	"context"
	"strings"

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

	idCol         = "id"
	taskIDCol     = "task_id"
	titleCol      = "title"
	urlCol        = "url"
	statusCol     = "status"
	weightCol     = "weight"
	uuidCol       = "uuid"
	parentUUIDCol = "parent_uuid"
)

var readColumns = []string{idCol, taskIDCol, titleCol, urlCol, statusCol, weightCol, uuidCol, parentUUIDCol}

type sourcePG struct {
	ID         int64   `db:"id"`
	TaskID     int64   `db:"task_id"`
	Title      string  `db:"title"`
	URL        string  `db:"url"`
	Status     string  `db:"status"`
	Weight     float64 `db:"weight"`
	UUID       string  `db:"uuid"`
	ParentUUID *string `db:"parent_uuid"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
}

func (s *Storage) Create(ctx context.Context, params []storage.ToCreateSource) ([]int64, error) {
	if len(params) == 0 {
		return nil, nil
	}

	q := pgSql.
		Insert(sourcesTbl).
		Columns(taskIDCol, titleCol, urlCol, weightCol, uuidCol, parentUUIDCol)

	for _, p := range params {
		q = q.Values(p.TaskID, p.Title, p.URL, p.Weight, p.UUID, p.ParentUUID)
	}

	sql, args := q.Suffix(returning(idCol)).MustSql()

	var out []int64

	err := s.db.SelectContext(ctx, &out, sql, args...)

	return out, err
}

func (s *Storage) GetByTaskID(ctx context.Context, taskID int64) ([]source.Source, error) {
	var out []sourcePG

	sql, args := pgSql.
		Select(readColumns...).
		From(sourcesTbl).
		Where(squirrel.Eq{taskIDCol: taskID}).
		MustSql()

	err := s.db.SelectContext(ctx, &out, sql, args...)

	return mapFromPGMany(out), err
}

func mapFromPGMany(sources []sourcePG) []source.Source {
	return lo.Map(sources, func(pg sourcePG, _ int) source.Source {
		return mapFromPG(pg)
	})
}

func mapFromPG(pg sourcePG) source.Source {
	return source.Source{
		ID:         pg.ID,
		TaskID:     pg.TaskID,
		Title:      pg.Title,
		URL:        pg.URL,
		Status:     source.Status(pg.Status),
		Weight:     pg.Weight,
		UUID:       pg.UUID,
		ParentUUID: pg.ParentUUID,
	}
}

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
