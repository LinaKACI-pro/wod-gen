package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
	"github.com/LinaKACI-pro/wod-gen/internal/handlers"
	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockWodGenerator struct {
	wod models.Wod
	err error
}

func (m *mockWodGenerator) Generate(ctx context.Context, level string, durationMin int, equipment []string, seed *string) (models.Wod, error) {
	if m.err != nil {
		return models.Wod{}, m.err
	}
	return m.wod, nil
}

type mockWodList struct {
	wods []models.Wod
	err  error
}

func (m *mockWodList) List(ctx context.Context, limit, offset int) ([]models.Wod, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.wods, nil
}

func TestGenerateWod_MissingBody(t *testing.T) {
	s := handlers.NewServer(&mockWodGenerator{}, &mockWodList{})

	resp, err := s.GenerateWod(context.Background(), handlers.GenerateWodRequestObject{Body: nil})
	require.NoError(t, err)

	r := resp.(*handlers.GenerateWod400JSONResponse)
	require.Equal(t, 400, r.Code)
	require.Equal(t, "missing body", r.Message)
}

func TestGenerateWod_Success(t *testing.T) {
	mockWod := models.Wod{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		Level:       "beginner",
		DurationMin: 20,
		Equipment:   []string{"rower"},
		Seed:        "seed123",
		Blocks:      []models.Block{{Name: "Run", Params: map[string]interface{}{"meters": 200}}},
	}

	s := handlers.NewServer(&mockWodGenerator{wod: mockWod}, &mockWodList{})

	body := handlers.GenerateWodJSONRequestBody{
		Level:       "beginner",
		DurationMin: 20,
	}
	resp, err := s.GenerateWod(context.Background(), handlers.GenerateWodRequestObject{Body: &body})
	require.NoError(t, err)

	r := resp.(*handlers.GenerateWod200JSONResponse)
	require.Equal(t, "beginner", string(r.Level))
	require.NotEmpty(t, r.Blocks)
}

func TestGenerateWod_ErrorKnown_InvalidData(t *testing.T) {
	s := handlers.NewServer(&mockWodGenerator{err: common.InvalidDataError{DataType: "level", Data: "bad"}}, &mockWodList{})

	body := handlers.GenerateWodJSONRequestBody{
		Level:       "wrong",
		DurationMin: 20,
	}
	resp, err := s.GenerateWod(context.Background(), handlers.GenerateWodRequestObject{Body: &body})
	require.NoError(t, err)

	r := resp.(*handlers.GenerateWod400JSONResponse)
	require.Equal(t, 400, r.Code)
	require.Contains(t, r.Message, "level") // message explicite
}

func TestGenerateWod_ErrorKnown_NoMoves(t *testing.T) {
	s := handlers.NewServer(&mockWodGenerator{err: common.ErrNoMoves}, &mockWodList{})

	body := handlers.GenerateWodJSONRequestBody{
		Level:       "beginner",
		DurationMin: 20,
	}
	resp, err := s.GenerateWod(context.Background(), handlers.GenerateWodRequestObject{Body: &body})
	require.NoError(t, err)

	r := resp.(*handlers.GenerateWod400JSONResponse)
	require.Equal(t, 400, r.Code)
	require.Equal(t, common.ErrNoMoves.Error(), r.Message)
}

func TestGenerateWod_ErrorUnknown_Internal(t *testing.T) {
	s := handlers.NewServer(&mockWodGenerator{err: errors.New("unexpected failure")}, &mockWodList{})

	body := handlers.GenerateWodJSONRequestBody{
		Level:       "beginner",
		DurationMin: 20,
	}
	resp, err := s.GenerateWod(context.Background(), handlers.GenerateWodRequestObject{Body: &body})
	require.NoError(t, err)

	r := resp.(*handlers.GenerateWod500JSONResponse)
	require.Equal(t, 500, r.Code)
	require.Equal(t, "internal server error", r.Message)
}

func TestListWods_Success(t *testing.T) {
	mockWod := models.Wod{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		Level:       "beginner",
		DurationMin: 20,
		Equipment:   []string{"rower"},
		Seed:        "seed123",
		Blocks:      []models.Block{{Name: "Run", Params: map[string]interface{}{"meters": 200}}},
	}
	s := handlers.NewServer(&mockWodGenerator{}, &mockWodList{wods: []models.Wod{mockWod}})

	resp, err := s.ListWods(context.Background(), handlers.ListWodsRequestObject{})
	require.NoError(t, err)

	r := resp.(*handlers.ListWods200JSONResponse)
	require.Len(t, *r.Wods, 1)
	require.Equal(t, "beginner", string((*r.Wods)[0].Level))
}

func TestListWods_ErrorFromRepo(t *testing.T) {
	s := handlers.NewServer(&mockWodGenerator{}, &mockWodList{err: errors.New("db fail")})

	resp, err := s.ListWods(context.Background(), handlers.ListWodsRequestObject{})
	require.NoError(t, err)

	r := resp.(*handlers.ListWods500JSONResponse)
	require.Equal(t, 500, r.Code)
}
