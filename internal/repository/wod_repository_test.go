package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/LinaKACI-pro/wod-gen/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newWod() models.Wod {
	return models.Wod{
		ID:          uuid.New(),
		Seed:        "abc",
		CreatedAt:   time.Now(),
		Level:       "beginner",
		DurationMin: 20,
		Equipment:   []string{"rower"},
		Blocks:      []models.Block{{Name: "Run", Params: map[string]interface{}{"meters": 200}}},
	}
}

func TestSaveWod_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()

	wod := newWod()
	mock.ExpectExec("INSERT INTO wods").
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := repository.NewWodRepository(db)
	got, err := repo.SaveWod(context.Background(), wod)

	require.NoError(t, err)
	require.Equal(t, wod.ID, got.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveWod_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()

	wod := newWod()
	mock.ExpectExec(`INSERT INTO wods`).
		WillReturnError(errors.New("db fail"))

	repo := repository.NewWodRepository(db)
	_, err := repo.SaveWod(context.Background(), wod)

	require.Error(t, err)
	require.Contains(t, err.Error(), "db.ExecContext")
}

func TestListWods_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()

	wod := newWod()
	blocks := `[{"name":"Run","params":{"meters":200}}]`

	rows := sqlmock.NewRows([]string{
		"id", "seed", "created_at", "level", "duration_min", "equipment", "blocks",
	}).AddRow(wod.ID, wod.Seed, wod.CreatedAt, wod.Level, wod.DurationMin, `{rower}`, blocks)

	mock.ExpectQuery("SELECT id, seed").
		WillReturnRows(rows)

	repo := repository.NewWodRepository(db)
	wods, err := repo.ListWods(context.Background(), 5, 0)

	require.NoError(t, err)
	require.Len(t, wods, 1)
	require.Equal(t, wod.Level, wods[0].Level)
}

func TestListWods_QueryError(t *testing.T) {
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery("SELECT id, seed").
		WillReturnError(errors.New("db fail"))

	repo := repository.NewWodRepository(db)
	_, err := repo.ListWods(context.Background(), 5, 0)

	require.Error(t, err)
	require.Contains(t, err.Error(), "db.QueryContext")
}
