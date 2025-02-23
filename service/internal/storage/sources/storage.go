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

	defaultWeight = 0
)

var readColumns = []string{idCol, taskIDCol, titleCol, urlCol, statusCol, weightCol, uuidCol, parentUUIDCol}

type sourcePG struct {
	id         int64   `db:"id"`
	taskID     int64   `db:"task_id"`
	title      string  `db:"title"`
	url        string  `db:"url"`
	status     string  `db:"status"`
	weight     float64 `db:"weight"`
	uuid       string  `db:"uuid"`
	parentUUID *string `db:"parent_uuid"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
}

func (s *Storage) Create(ctx context.Context, params []storage.ToCreateSource) error {
	if len(params) == 0 {
		return nil
	}

	q := pgSql.
		Insert(sourcesTbl).
		Columns(taskIDCol, titleCol, urlCol, weightCol, uuidCol, parentUUIDCol)

	for _, p := range params {
		q.Values(p.TaskID, p.Title, p.URL, p.Weight, p.UUID, p.ParentUUID)
	}

	sql, args := q.MustSql()

	_, err := s.db.ExecContext(ctx, sql, args...)

	return err
}

func (s *Storage) GetByTaskID(ctx context.Context, taskID int64) ([]source.Source, error) {
	sql, args := pgSql.
		Select(readColumns...).
		From(sourcesTbl).
		Where(squirrel.Eq{taskIDCol: taskID}).
		MustSql()

	var out []sourcePG
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
		ID:         pg.id,
		TaskID:     pg.taskID,
		Title:      pg.title,
		URL:        pg.url,
		Status:     source.Status(pg.status),
		Weight:     pg.weight,
		UUID:       pg.uuid,
		ParentUUID: pg.parentUUID,
	}
}

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
