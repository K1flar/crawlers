package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/models/task"
	task_model "github.com/K1flar/crawlers/internal/models/task"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var _ storage.Tasks = (*Storage)(nil)

type Storage struct {
	db *sqlx.DB
}

var pgSql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	tasksTbl = "tasks"

	idCol         = "id"
	queryCol      = "query"
	statusCol     = "status"
	createdAtCol  = "created_at"
	updatedAtCol  = "updated_at"
	depthLevelCol = "depth_level"
	minWeightCol  = "min_weight"

	defaultDepthLevelCol = 3
	defaultMinWeightCol  = 0
)

var readColumns = []string{idCol, queryCol, statusCol, createdAtCol, updatedAtCol, depthLevelCol, minWeightCol}

type taskPG struct {
	ID         int64     `db:"id"`
	Query      string    `db:"query"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	DepthLevel int       `db:"depth_level"`
	MinWeight  float64   `db:"min_weight"`
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
		).
		Values(
			params.Query,
			task_model.StatusCreated,
			now,
			now,
			defaultDepthLevelCol,
			defaultMinWeightCol,
		).
		Suffix(returning(idCol)).
		MustSql()

	err := s.db.GetContext(ctx, &id, sql, args...)

	return id, err
}

func mapFromPG(pg taskPG) task.Task {
	return task.Task{
		ID:         pg.ID,
		Query:      pg.Query,
		Status:     task.Status(pg.Status),
		CreatedAt:  pg.CreatedAt,
		UpdatedAt:  pg.UpdatedAt,
		DepthLevel: pg.DepthLevel,
		MinWeight:  pg.MinWeight,
	}
}

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
