package core

import (
	"context"
	"reflect"
	"testing"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
	"github.com/stretchr/testify/require"
)

const miniCatalogYAML = `
moves:
  - name: Run
    needs_one_of: []
    weight: 1
    ranges:
      beginner: { meters: [200, 400] }
      intermediate: { meters: [400, 800] }
      advanced: { meters: [800, 1200] }
  - name: Push-ups
    needs_one_of: []
    weight: 1
    ranges:
      beginner: { reps: [5, 10] }
      intermediate: { reps: [10, 20] }
      advanced: { reps: [20, 30] }
  - name: DB Snatch
    needs_one_of: ["dumbbell"]
    weight: 1
    ranges:
      beginner: { reps: [6, 10] }
      intermediate: { reps: [8, 12] }
      advanced: { reps: [10, 16] }
`

func TestGenerate_DeterministicWithSeed(t *testing.T) {
	t.Parallel()
	cat, err := NewCatalog([]byte(miniCatalogYAML))
	if err != nil {
		t.Fatalf("NewCatalog error: %v", err)
	}

	seed := "abc123-seed"
	level := "Intermediate"
	duration := 45
	equip := []string{"dumbbell"}

	w1, err := cat.Generate(context.Background(), level, duration, equip, &seed)
	if err != nil {
		t.Fatalf("Generate #1 error: %v", err)
	}

	w2, err := cat.Generate(context.Background(), level, duration, equip, &seed)
	if err != nil {
		t.Fatalf("Generate #2 error: %v", err)
	}

	if !reflect.DeepEqual(w1.Blocks, w2.Blocks) {
		t.Fatalf("expected identical blocks with same seed, got different:\n%v\nvs\n%v", w1.Blocks, w2.Blocks)
	}

	if w1.Level != "intermediate" {
		t.Fatalf("expected level 'intermediate', got %q", w1.Level)
	}
	if w1.Seed != seed {
		t.Fatalf("expected seed %q, got %q", seed, w1.Seed)
	}
	if w1.DurationMin != duration {
		t.Fatalf("expected duration %d, got %d", duration, w1.DurationMin)
	}
}

func TestNewCatalog_InvalidYAML(t *testing.T) {
	raw := []byte(`:: invalid yaml ::`)
	c, err := NewCatalog(raw)
	require.Error(t, err)
	require.Nil(t, c)
}

func TestGenerate_InvalidLevel(t *testing.T) {
	raw := []byte(`
moves:
  - name: Push-ups
    weight: 1
`)
	c, err := NewCatalog(raw)
	require.NoError(t, err)

	_, err = c.Generate(context.Background(), "invalid-level", 30, nil, nil)
	require.Error(t, err)
	require.IsType(t, common.InvalidDataError{}, err)
}

func TestGenerate_DurationOutOfRange(t *testing.T) {
	raw := []byte(`
moves:
  - name: Squats
    weight: 1
`)
	c, err := NewCatalog(raw)
	require.NoError(t, err)

	_, err = c.Generate(context.Background(), Beginner, 5, nil, nil) // trop petit
	require.Error(t, err)
	require.Equal(t, common.ErrDuration, err)

	_, err = c.Generate(context.Background(), Beginner, 200, nil, nil) // trop grand
	require.Error(t, err)
	require.Equal(t, common.ErrDuration, err)
}

func TestGenerate_EmptyCatalog(t *testing.T) {
	c := &Catalog{Moves: []move{}}
	_, err := c.Generate(context.Background(), Beginner, 30, nil, nil)
	require.Error(t, err)
	require.Equal(t, common.ErrEmptyCatalog, err)
}

func TestGenerate_NoMovesAvailable(t *testing.T) {
	raw := []byte(`
moves:
  - name: Bike
    needs_one_of: ["bike"]
    weight: 1
`)
	c, err := NewCatalog(raw)
	require.NoError(t, err)

	// On demande un équipement inexistant → aucun move dispo
	_, err = c.Generate(context.Background(), Beginner, 30, []string{"rower"}, nil)
	require.Error(t, err)
	require.Equal(t, common.ErrNoMoves, err)
}

func TestGenerate_Success(t *testing.T) {
	raw := []byte(`
moves:
  - name: Push-ups
    weight: 1
    ranges:
      beginner:
        reps: [5, 5]
`)
	c, err := NewCatalog(raw)
	require.NoError(t, err)

	wod, err := c.Generate(context.Background(), Beginner, 20, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, wod.Blocks)
	require.Equal(t, Beginner, wod.Level)
	require.Equal(t, 20, wod.DurationMin)
	require.NotEmpty(t, wod.Seed)
}
