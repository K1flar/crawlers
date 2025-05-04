package task_sources

import (
	"context"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

var _ storage.TaskSources = (*Storage)(nil)

type Storage struct {
	db *sqlx.DB
}

var pgSql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	tasksSourcesTbl = "tasks_x_sources"
	sourcesTbl      = "sources"

	taskIDCol         = "task_id"
	launchIDCol       = "launch_id"
	sourceIDCol       = "source_id"
	parentSourceIDCol = "parent_source_id"
	weightCol         = "weight"
)

type sourcePG struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Title     string    `db:"title"`
	URL       string    `db:"url"`
	Status    string    `db:"status"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
}

func (s *Storage) GetByTaskID(ctx context.Context, taskID int64) ([]source.Source, error) {
	var res []sourcePG

	sql, args := pgSql.
		Select("s.id", "s.created_at", "s.updated_at", "s.title", "s.url", "s.status").
		From("sources s").
		Join("tasks_x_sources txs ON s.id = txs.source_id").
		Where(squirrel.Eq{"txs.task_id": taskID}).
		Distinct().
		MustSql()

	err := s.db.SelectContext(ctx, &res, sql, args...)

	return mapFromPGManySources(res), err
}

func (s *Storage) Create(ctx context.Context, params []storage.ToCreateTaskSource) error {
	if len(params) == 0 {
		return nil
	}

	q := pgSql.
		Insert(tasksSourcesTbl).
		Columns(taskIDCol, launchIDCol, sourceIDCol, parentSourceIDCol, weightCol)

	for _, p := range params {
		q = q.Values(p.TaskID, p.LaunchID, p.SourceID, p.ParentSourceID, p.Weight)
	}

	sql, args := q.MustSql()

	_, err := s.db.ExecContext(ctx, sql, args...)

	return err
}

func mapFromPGManySources(sources []sourcePG) []source.Source {
	return lo.Map(sources, func(pg sourcePG, _ int) source.Source {
		return mapFromPGSource(pg)
	})
}

func mapFromPGSource(pg sourcePG) source.Source {
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
