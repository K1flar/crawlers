package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

var _ storage.Tasks = (*Storage)(nil)

type Storage struct {
	db *sqlx.DB
}

var pgSql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	tasksTbl = "tasks"

	idCol                     = "id"
	queryCol                  = "query"
	statusCol                 = "status"
	createdAtCol              = "created_at"
	updatedAtCol              = "updated_at"
	processedAtCol            = "processed_at"
	depthLevelCol             = "depth_level"
	minWeightCol              = "min_weight"
	maxSourcesCol             = "max_sources"
	maxNeighboursForSourceCol = "max_neighbours_for_source"
)

var readColumns = []string{
	idCol,
	queryCol,
	statusCol,
	createdAtCol,
	updatedAtCol,
	processedAtCol,
	depthLevelCol,
	minWeightCol,
	maxSourcesCol,
	maxNeighboursForSourceCol,
}

type taskPG struct {
	ID                     int64      `db:"id"`
	Query                  string     `db:"query"`
	Status                 string     `db:"status"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
	ProcessedAt            *time.Time `db:"processed_at"`
	DepthLevel             int        `db:"depth_level"`
	MinWeight              float64    `db:"min_weight"`
	MaxSources             int64      `db:"max_sources"`
	MaxNeighboursForSource int64      `db:"max_neighbours_for_source"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
}

func (s *Storage) GetByID(ctx context.Context, id int64) (task.Task, error) {
	var task taskPG

	sql, args := pgSql.
		Select(readColumns...).
		From(tasksTbl).
		Where(squirrel.Eq{idCol: id}).
		MustSql()

	err := s.db.GetContext(ctx, &task, sql, args...)

	return mapFromPG(task), err
}

func (s *Storage) FindInStatuses(ctx context.Context, statuses []task.Status) ([]task.Task, error) {
	var tasks []taskPG

	sql, args := pgSql.
		Select(readColumns...).
		From(tasksTbl).
		Where(squirrel.Eq{statusCol: statuses}).
		MustSql()

	err := s.db.SelectContext(ctx, &tasks, sql, args...)

	return mapFromPgMany(tasks), err
}

func (s *Storage) Create(ctx context.Context, params storage.ToCreateTask) (int64, error) {
	var id int64

	now := time.Now()

	sql, args := pgSql.
		Insert(tasksTbl).
		Columns(
			queryCol,
			statusCol,
			createdAtCol,
			updatedAtCol,
			depthLevelCol,
			minWeightCol,
			maxSourcesCol,
			maxNeighboursForSourceCol,
		).
		Values(
			params.Query,
			task.StatusCreated,
			now,
			now,
			params.DepthLevel,
			params.MinWeight,
			params.MaxSources,
			params.MaxNeighboursForSource,
		).
		Suffix(returning(idCol)).
		MustSql()

	err := s.db.GetContext(ctx, &id, sql, args...)

	return id, err
}

func (s *Storage) SetStatus(ctx context.Context, id int64, status task.Status) error {
	sql, args := pgSql.
		Update(tasksTbl).
		Set(statusCol, status).
		Set(updatedAtCol, time.Now()).
		Where(squirrel.Eq{idCol: id}).
		MustSql()

	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return business_errors.EntityNotFound
	}

	return nil
}

func mapFromPG(pg taskPG) task.Task {
	return task.Task{
		ID:                     pg.ID,
		Query:                  pg.Query,
		Status:                 task.Status(pg.Status),
		CreatedAt:              pg.CreatedAt,
		UpdatedAt:              pg.UpdatedAt,
		ProcessedAt:            pg.ProcessedAt,
		DepthLevel:             pg.DepthLevel,
		MinWeight:              pg.MinWeight,
		MaxSources:             pg.MaxSources,
		MaxNeighboursForSource: pg.MaxNeighboursForSource,
	}
}

func mapFromPgMany(pgs []taskPG) []task.Task {
	return lo.Map(pgs, func(pg taskPG, _ int) task.Task {
		return mapFromPG(pg)
	})
}

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
