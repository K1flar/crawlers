package sources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/models/source"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/K1flar/crawlers/internal/utils"
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
	sourcesTbl  = "sources"
	tasksTbl    = "tasks"
	launchesTbl = "launches"

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

type taskSourcePG struct {
	ID       int64   `db:"id"`
	URL      string  `db:"url"`
	Title    string  `db:"title"`
	Weight   float64 `db:"weight"`
	ParentID *int64  `db:"parent_source_id"`
}

func (s *Storage) GetByTaskID(ctx context.Context, taskID int64) ([]source.ForTask, error) {
	var res []taskSourcePG

	subSql := squirrel.Expr("txs.launch_id = (SELECT MAX(id) FROM launches WHERE task_id = ?)", taskID)

	sql, args := pgSql.
		Select("s.id", "s.title", "s.url", "txs.weight", "txs.parent_source_id").
		From("sources s").
		Join("tasks_x_sources txs ON s.id = txs.source_id").
		Where(squirrel.Eq{"txs.task_id": taskID}).
		Where(subSql).
		MustSql()

	err := s.db.SelectContext(ctx, &res, sql, args...)

	return lo.Map(res, func(pg taskSourcePG, _ int) source.ForTask {
		return source.ForTask{
			ID:       pg.ID,
			URL:      pg.URL,
			Title:    pg.Title,
			Weight:   pg.Weight,
			ParentID: pg.ParentID,
		}
	}), err
}

type sourceForProtocolPG struct {
	TaskID          int64      `db:"task_id"`
	Query           string     `db:"query"`
	SourceID        int64      `db:"source_id"`
	Title           string     `db:"title"`
	URL             string     `db:"url"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	SourceStatus    string     `db:"source_status"`
	LaunchID        int64      `db:"launch_id"`
	LaunchNumber    int64      `db:"launch_number"`
	StartedAt       time.Time  `db:"started_at"`
	FinishedAt      *time.Time `db:"finished_at"`
	LaunchStatus    string     `db:"launch_status"`
	LaunchErrorSlug *string    `db:"error"`
}

func (s *Storage) GetForProtocol(ctx context.Context, filter storage.FilterForProtocol) ([]source.ForProtocol, error) {
	var res []sourceForProtocolPG

	q := pgSql.
		Select("t.id as task_id", "t.query",
			"s.id as source_id", "s.title", "s.url", "s.created_at", "s.updated_at", "s.status as source_status",
			"l.id as launch_id", "l.number as launch_number", "l.started_at",
			"finished_at", "l.status as launch_status", "l.error").
		From("sources s").
		Join("tasks_x_sources txs ON s.id = txs.source_id").
		Join("tasks t ON t.id = txs.task_id").
		Join("launches l ON t.id = l.task_id")

	if filter.TaskID != nil {
		q = q.Where(squirrel.Eq{"t.id": filter.TaskID})
	}

	if filter.Query != nil {
		q = q.Where(squirrel.Like{"t.query": fmt.Sprintf("%%%s%%", *filter.Query)})
	}

	if filter.SourceID != nil {
		q = q.Where(squirrel.Eq{"s.id": filter.SourceID})
	}

	if filter.Title != nil {
		q = q.Where(squirrel.Like{"s.title": fmt.Sprintf("%%%s%%", *filter.Title)})
	}

	if filter.SourceStatus != nil {
		q = q.Where(squirrel.Like{"s.status": fmt.Sprintf("%%%s%%", *filter.SourceStatus)})
	}

	sql, args := q.Limit(uint64(filter.Limit)).Offset(uint64(filter.Offset)).MustSql()

	err := s.db.SelectContext(ctx, &res, sql, args...)

	return lo.Map(res, func(s sourceForProtocolPG, _ int) source.ForProtocol {
		var duration *time.Duration
		if s.FinishedAt != nil {
			duration = utils.Ptr(s.FinishedAt.Sub(s.StartedAt))
		}

		return source.ForProtocol{
			TaskID:          s.TaskID,
			Query:           s.Query,
			SourceID:        s.SourceID,
			Title:           s.Title,
			URL:             s.URL,
			CreatedAt:       s.CreatedAt,
			UpdatedAt:       s.UpdatedAt,
			SourceStatus:    source.Status(s.SourceStatus),
			LaunchID:        s.LaunchID,
			LaunchNumber:    s.LaunchNumber,
			StartedAt:       s.StartedAt,
			Duration:        duration,
			LaunchStatus:    launch.Status(s.LaunchStatus),
			LaunchErrorSlug: s.LaunchErrorSlug,
		}
	}), err
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
