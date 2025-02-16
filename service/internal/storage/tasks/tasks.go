package tasks

import (
	"context"
	"strings"
	"time"

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

type taskPG struct {
	id         int64     `db:"id"`
	query      string    `db:"query"`
	status     string    `db:"status"`
	createdAt  time.Time `db:"created_at"`
	updatedAt  time.Time `db:"updated_at"`
	depthLevel int       `db:"depth_level"`
	minWeight  float64   `db:"min_weight"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
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

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
