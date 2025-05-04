package launches

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/K1flar/crawlers/internal/business_errors"
	"github.com/K1flar/crawlers/internal/models/launch"
	"github.com/K1flar/crawlers/internal/storage"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var _ storage.Launches = (*Storage)(nil)

type Storage struct {
	db *sqlx.DB
}

var pgSql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	launchesTbl = "launches"

	idCol            = "id"
	numberCol        = "number"
	taskIDCol        = "task_id"
	startedAtCol     = "started_at"
	finishedAtCol    = "finished_at"
	sourcesViewedCol = "sources_viewed"
	statusCol        = "status"
	errorCol         = "error"
)

var readColumns = []string{idCol, numberCol, taskIDCol, startedAtCol, finishedAtCol, sourcesViewedCol, statusCol, errorCol}

type launchPG struct {
	ID            int64      `db:"id"`
	Number        int64      `db:"number"`
	TaskID        int64      `db:"task_id"`
	StartedAt     time.Time  `db:"started_at"`
	FinishedAt    *time.Time `db:"finished_at"`
	SourcesViewed int64      `db:"sources_viewed"`
	Status        string     `db:"status"`
	Error         *string    `db:"error"`
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db}
}

func (s *Storage) Create(ctx context.Context, params storage.ToCreateLaunch) (int64, error) {
	var nextNumber int64

	err := pgSql.Select("COALESCE(MAX(number), 0) + 1").
		From("launches").
		Where(squirrel.Eq{"task_id": params.TaskID}).
		RunWith(s.db).
		ScanContext(ctx, &nextNumber)
	if err != nil {
		return 0, err
	}

	query, args := pgSql.
		Insert(launchesTbl).
		Columns(
			numberCol,
			taskIDCol,
			startedAtCol,
			sourcesViewedCol,
			statusCol,
		).
		Values(
			nextNumber,
			params.TaskID,
			params.StartedAt,
			0,
			launch.StatusInProgress,
		).
		Suffix(returning(idCol)).
		MustSql()

	var id int64
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return id, err
}

func (s *Storage) Finish(ctx context.Context, params storage.ToFinishLaunch) error {
	sql, args := pgSql.
		Update(launchesTbl).
		SetMap(map[string]any{
			finishedAtCol:    params.FinishedAt,
			sourcesViewedCol: params.SourcesViewed,
			statusCol:        launch.StatusFinished,
			errorCol:         params.Error,
		}).
		Where(squirrel.Eq{idCol: params.ID}).
		MustSql()

	res, err := s.db.ExecContext(ctx, sql, args...)

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("failed to finish launch: %w", business_errors.EntityNotFound)
	}

	return nil
}

func (s *Storage) Get(ctx context.Context, id int64) (launch.Launch, error) {
	var res launchPG

	sql, args := pgSql.
		Select(readColumns...).
		From(launchesTbl).
		Where(squirrel.Eq{idCol: id}).
		MustSql()

	err := s.db.GetContext(ctx, &res, sql, args...)

	return mapFromPG(res), err
}

func mapFromPG(pg launchPG) launch.Launch {
	return launch.Launch{
		ID:            pg.ID,
		Number:        pg.Number,
		TaskID:        pg.TaskID,
		StartedAt:     pg.StartedAt,
		FinishedAt:    pg.FinishedAt,
		SourcesViewed: pg.SourcesViewed,
		Status:        launch.Status(pg.Status),
		Error:         (*launch.ErrorSlug)(pg.Error),
	}
}

func returning(cols ...string) string {
	return "returning " + strings.Join(cols, ", ")
}
