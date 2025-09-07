package core

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
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
	cat, err := NewCatalog([]byte(miniCatalogYAML))
	if err != nil {
		t.Fatalf("NewCatalog error: %v", err)
	}

	seed := "abc123-seed"
	level := "Intermediate"
	duration := 45
	equip := []string{"dumbbell"}

	// 1ère génération
	w1, err := cat.Generate(context.Background(), level, duration, equip, &seed)
	if err != nil {
		t.Fatalf("Generate #1 error: %v", err)
	}

	// 2ème génération (mêmes inputs)
	w2, err := cat.Generate(context.Background(), level, duration, equip, &seed)
	if err != nil {
		t.Fatalf("Generate #2 error: %v", err)
	}

	// mêmes blocs attendus à seed identique
	if !reflect.DeepEqual(w1.Blocks, w2.Blocks) {
		t.Fatalf("expected identical blocks with same seed, got different:\n%v\nvs\n%v", w1.Blocks, w2.Blocks)
	}

	// vérifs légères : level normalisé, seed renvoyée telle quelle, durée respectée
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

func TestGenerate_InvalidLevel(t *testing.T) {
	cat, err := NewCatalog([]byte(miniCatalogYAML))
	if err != nil {
		t.Fatalf("NewCatalog error: %v", err)
	}

	_, err = cat.Generate(context.Background(), "novice", 30, nil, nil)
	if err == nil {
		t.Fatalf("expected error for invalid level, got nil")
	}

	var inv common.InvalidDataError
	if !errors.As(err, &inv) {
		t.Fatalf("expected InvalidDataError, got %T: %v", err, err)
	}
}
