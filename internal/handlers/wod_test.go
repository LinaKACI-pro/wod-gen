package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeGenerator struct {
	wod models.Wod
	err error
}

func (f *fakeGenerator) Generate(ctx context.Context, level string, durationMin int, equipment []string, seed *string) (models.Wod, error) {
	_, _, _, _, _ = ctx, level, durationMin, equipment, seed
	return f.wod, f.err
}

func TestGenerateWod_Success(t *testing.T) {
	now := time.Now().UTC()
	expected := models.Wod{
		ID:          uuid.New(),
		CreatedAt:   now,
		Level:       "beginner",
		DurationMin: 30,
		Equipment:   []string{"rower"},
		Seed:        "abc123",
		Blocks: []models.Block{
			{Name: "Push-ups", Params: map[string]interface{}{"reps": 10}},
		},
	}

	server := NewServer(&fakeGenerator{wod: expected})

	req := GenerateWodRequestObject{
		Body: &GenerateWodJSONRequestBody{
			Level:       GenerateWodParamsLevel(expected.Level),
			DurationMin: expected.DurationMin,
			Equipment:   &expected.Equipment,
			Seed:        &expected.Seed,
		},
	}

	resp, err := server.GenerateWod(context.Background(), req)
	require.NoError(t, err)

	okResp, ok := resp.(*GenerateWod200JSONResponse)
	require.True(t, ok)

	require.Equal(t, expected.DurationMin, okResp.DurationMin)
	require.Equal(t, expected.Level, string(okResp.Level))
	require.Equal(t, expected.Seed, okResp.Seed)
	require.Len(t, okResp.Blocks, 1)
	require.Equal(t, "Push-ups", *okResp.Blocks[0].Name)
}

func TestGenerateWod_Error(t *testing.T) {
	server := NewServer(&fakeGenerator{err: errors.New("boom")})

	req := GenerateWodRequestObject{
		Body: &GenerateWodJSONRequestBody{
			Level:       "beginner",
			DurationMin: 20,
		},
	}

	resp, err := server.GenerateWod(context.Background(), req)
	require.NoError(t, err)

	errResp, ok := resp.(*GenerateWod400JSONResponse)
	require.True(t, ok)
	require.Equal(t, 400, errResp.Code)
	require.Contains(t, errResp.Message, "invalid request")
}

func TestGenerateWod_MissingBody(t *testing.T) {
	server := NewServer(&fakeGenerator{})

	req := GenerateWodRequestObject{Body: nil}
	resp, err := server.GenerateWod(context.Background(), req)
	require.NoError(t, err)

	errResp, ok := resp.(*GenerateWod400JSONResponse)
	require.True(t, ok)
	require.Equal(t, 400, errResp.Code)
}
