package core

import (
	"context"
	"errors"
	"testing"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
	"github.com/LinaKACI-pro/wod-gen/internal/core/catalog"
	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/stretchr/testify/require"
)

type mockWodRepo struct {
	saved models.Wod
	err   error
}

func (m *mockWodRepo) SaveWod(ctx context.Context, w models.Wod) (models.Wod, error) {
	if m.err != nil {
		return models.Wod{}, m.err
	}
	m.saved = w
	return w, nil
}

func (m *mockWodRepo) ListWods(ctx context.Context, limit, offset int) ([]models.Wod, error) {
	return []models.Wod{m.saved}, nil
}

func TestValidateInfo_InvalidLevel(t *testing.T) {
	_, _, err := validateInfo("expert", 30, nil, []catalog.Move{{Name: "Run"}})
	require.Error(t, err)
}

func TestValidateInfo_InvalidDuration(t *testing.T) {
	_, _, err := validateInfo("beginner", 5, nil, []catalog.Move{{Name: "Run"}})
	require.Error(t, err)
}

func TestValidateInfo_EmptyCatalog(t *testing.T) {
	_, _, err := validateInfo("beginner", 30, nil, []catalog.Move{})
	require.Error(t, err)
}

func TestGenerate_Success(t *testing.T) {
	moves := []catalog.Move{
		{
			Name:   "Run",
			Weight: 1,
			Ranges: map[string]map[string]catalog.Rng{
				"beginner": {"meters": {100, 200}},
			},
		},
	}

	repo := &mockWodRepo{}
	gen := NewWodGenerator(&catalog.Catalog{Moves: moves}, repo)

	wod, err := gen.Generate(context.Background(), "beginner", 30, []string{}, nil)
	require.NoError(t, err)
	require.Equal(t, "beginner", wod.Level)
	require.NotEmpty(t, wod.Blocks)
	require.Equal(t, repo.saved.ID, wod.ID)
}

func TestGenerate_SaveError(t *testing.T) {
	moves := []catalog.Move{
		{
			Name:   "Run",
			Weight: 1,
			Ranges: map[string]map[string]catalog.Rng{
				"beginner": {"meters": {100, 200}},
			},
		},
	}

	repo := &mockWodRepo{err: errors.New("db down")}
	gen := NewWodGenerator(&catalog.Catalog{Moves: moves}, repo)

	_, err := gen.Generate(context.Background(), "beginner", 30, []string{}, nil)
	require.Error(t, err)
}

func TestBuildWod_ErrNoMoves(t *testing.T) {
	moves := []catalog.Move{}
	_, err := buildWod("beginner", 20, []string{}, "seed", moves)
	require.ErrorIs(t, err, common.ErrNoMoves)
}
